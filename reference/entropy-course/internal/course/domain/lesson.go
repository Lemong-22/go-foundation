package domain

import (
	"sort"
	"strings"
	"time"
)

type Lesson struct {
	id        LessonID
	courseID  CourseID
	title     string
	blocks    []ContentBlock
	order     LessonOrder
	createdAt time.Time
	updatedAt time.Time
}

func NewLesson(
	id LessonID,
	courseID CourseID,
	title string,
	blocks []ContentBlock,
	order LessonOrder,
	now time.Time,
) (Lesson, error) {
	normalizedTitle, err := normalizeTitle(title)
	if err != nil {
		return Lesson{}, err
	}

	normalizedBlocks, err := normalizeLessonBlocks(blocks)
	if err != nil {
		return Lesson{}, err
	}

	return Lesson{
		id:        id,
		courseID:  courseID,
		title:     normalizedTitle,
		blocks:    normalizedBlocks,
		order:     order,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func RestoreLesson(
	id LessonID,
	courseID CourseID,
	title string,
	blocks []ContentBlock,
	order LessonOrder,
	createdAt time.Time,
	updatedAt time.Time,
) (Lesson, error) {
	normalizedTitle, err := normalizeTitle(title)
	if err != nil {
		return Lesson{}, err
	}

	normalizedBlocks, err := normalizeLessonBlocks(blocks)
	if err != nil {
		return Lesson{}, err
	}

	return Lesson{
		id:        id,
		courseID:  courseID,
		title:     normalizedTitle,
		blocks:    normalizedBlocks,
		order:     order,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}, nil
}

func (lesson Lesson) ID() LessonID {
	return lesson.id
}

func (lesson Lesson) CourseID() CourseID {
	return lesson.courseID
}

func (lesson Lesson) Title() string {
	return lesson.title
}

func (lesson Lesson) Blocks() []ContentBlock {
	blocks := make([]ContentBlock, len(lesson.blocks))
	copy(blocks, lesson.blocks)

	return blocks
}

func (lesson Lesson) Block(id BlockID) (ContentBlock, error) {
	for _, block := range lesson.blocks {
		if block.ID() == id {
			return block, nil
		}
	}

	return ContentBlock{}, ErrNotFound
}

func (lesson Lesson) Content() string {
	markdownBlocks := make([]string, 0, len(lesson.blocks))
	for _, block := range lesson.blocks {
		if !block.Kind().IsText() {
			continue
		}

		body, ok := block.Body().(TextBody)
		if !ok {
			continue
		}

		markdownBlocks = append(markdownBlocks, body.Markdown)
	}

	return strings.Join(markdownBlocks, "\n\n")
}

func (lesson Lesson) Order() LessonOrder {
	return lesson.order
}

func (lesson Lesson) CreatedAt() time.Time {
	return lesson.createdAt
}

func (lesson Lesson) UpdatedAt() time.Time {
	return lesson.updatedAt
}

func (lesson *Lesson) Rename(title string, now time.Time) error {
	normalizedTitle, err := normalizeTitle(title)
	if err != nil {
		return err
	}

	lesson.title = normalizedTitle
	lesson.touch(now)

	return nil
}

func (lesson *Lesson) ChangeContent(content string, now time.Time) {
	blockID, _ := NewBlockID(lesson.id.String())
	position := BlockPosition{value: 0}
	block, _ := NewTextBlock(blockID, position, content)

	lesson.blocks = []ContentBlock{block}
	lesson.touch(now)
}

func (lesson *Lesson) MoveTo(order LessonOrder, now time.Time) {
	lesson.order = order
	lesson.touch(now)
}

func (lesson *Lesson) AddBlock(block ContentBlock, now time.Time) error {
	if lesson.hasBlockID(block.ID()) {
		return NewValidationError("block_id", "must be unique within the lesson")
	}
	if block.Position().Int() > len(lesson.blocks) {
		return NewValidationError("position", "must be less than or equal to block count")
	}

	blocks := lesson.Blocks()
	for i := range blocks {
		if blocks[i].Position().Int() >= block.Position().Int() {
			blocks[i].MoveTo(BlockPosition{value: blocks[i].Position().Int() + 1})
		}
	}
	blocks = append(blocks, block)
	lesson.blocks = sortBlocksByPosition(blocks)
	lesson.touch(now)

	return nil
}

func (lesson *Lesson) UpdateBlock(id BlockID, body ContentBody, now time.Time) error {
	blocks := lesson.Blocks()
	for i := range blocks {
		if blocks[i].ID() != id {
			continue
		}

		if err := blocks[i].ChangeBody(body); err != nil {
			return err
		}

		lesson.blocks = blocks
		lesson.touch(now)
		return nil
	}

	return ErrNotFound
}

func (lesson *Lesson) RemoveBlock(id BlockID, now time.Time) error {
	blocks := make([]ContentBlock, 0, len(lesson.blocks))
	removed := false

	for _, block := range lesson.blocks {
		if block.ID() == id {
			removed = true
			continue
		}

		block.MoveTo(BlockPosition{value: len(blocks)})
		blocks = append(blocks, block)
	}

	if !removed {
		return ErrNotFound
	}

	lesson.blocks = blocks
	lesson.touch(now)

	return nil
}

func (lesson *Lesson) ReorderBlocks(order []BlockPlacement, now time.Time) error {
	if len(order) != len(lesson.blocks) {
		return NewValidationError("order", "must include every block exactly once")
	}

	current := make(map[string]ContentBlock, len(lesson.blocks))
	for _, block := range lesson.blocks {
		current[block.ID().String()] = block
	}

	usedBlocks := make(map[string]struct{}, len(order))
	usedPositions := make(map[int]struct{}, len(order))
	positions := make(map[string]BlockPosition, len(order))

	for _, placement := range order {
		id := placement.BlockID.String()
		if _, exists := current[id]; !exists {
			return NewValidationError("block_id", "must belong to the lesson")
		}
		if _, exists := usedBlocks[id]; exists {
			return NewValidationError("block_id", "must be unique")
		}

		position := placement.Position.Int()
		if position >= len(order) {
			return NewValidationError("position", "must be contiguous from zero")
		}
		if _, exists := usedPositions[position]; exists {
			return NewValidationError("position", "must be unique")
		}

		usedBlocks[id] = struct{}{}
		usedPositions[position] = struct{}{}
		positions[id] = placement.Position
	}

	blocks := make([]ContentBlock, 0, len(lesson.blocks))
	for _, block := range lesson.blocks {
		block.MoveTo(positions[block.ID().String()])
		blocks = append(blocks, block)
	}

	lesson.blocks = sortBlocksByPosition(blocks)
	lesson.touch(now)

	return nil
}

func (lesson *Lesson) touch(now time.Time) {
	lesson.updatedAt = mutationTime(lesson.createdAt, now)
}

func (lesson Lesson) hasBlockID(id BlockID) bool {
	for _, block := range lesson.blocks {
		if block.ID() == id {
			return true
		}
	}

	return false
}

func normalizeLessonBlocks(blocks []ContentBlock) ([]ContentBlock, error) {
	normalized := make([]ContentBlock, len(blocks))
	copy(normalized, blocks)

	ids := make(map[string]struct{}, len(normalized))
	positions := make(map[int]struct{}, len(normalized))
	for _, block := range normalized {
		id := block.ID().String()
		if _, exists := ids[id]; exists {
			return nil, NewValidationError("block_id", "must be unique within the lesson")
		}
		ids[id] = struct{}{}

		position := block.Position().Int()
		if _, exists := positions[position]; exists {
			return nil, NewValidationError("position", "must be unique")
		}
		positions[position] = struct{}{}
	}

	normalized = sortBlocksByPosition(normalized)
	for i, block := range normalized {
		if block.Position().Int() != i {
			return nil, NewValidationError("position", "must be contiguous from zero")
		}
	}

	return normalized, nil
}

func sortBlocksByPosition(blocks []ContentBlock) []ContentBlock {
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Position().Int() < blocks[j].Position().Int()
	})

	return blocks
}
