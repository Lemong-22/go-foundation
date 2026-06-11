package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/spf13/cobra"
)

func newLessonBlockCommand(context lessonCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "block",
		Short: "Manage lesson blocks",
	}

	command.AddCommand(
		newLessonBlockAddCommand(context),
		newLessonBlockListCommand(context),
		newLessonBlockGetCommand(context),
		newLessonBlockUpdateCommand(context),
		newLessonBlockRemoveCommand(context),
		newLessonBlockReorderCommand(context),
	)

	return command
}

func newLessonBlockAddCommand(context lessonCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "add",
		Short: "Add a lesson block",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			input, err := lessonBlockAddInput(command)
			if err != nil {
				return err
			}

			out, err := context.service.AddLessonBlock(input)
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderCreatedBlock(out.ID)
		},
	}

	flags := command.Flags()
	flags.String("lesson-id", "", "lesson id")
	flags.String("kind", "", "block kind")
	flags.String("text", "", "text markdown")
	flags.String("video-provider", "", "video provider")
	flags.String("video-locator", "", "video locator")
	flags.String("video-caption", "", "video caption")
	flags.String("quiz-id", "", "quiz id")
	flags.String("practice-id", "", "practice id")
	flags.Int("position", 0, "block position")

	return command
}

func lessonBlockAddInput(command *cobra.Command) (core.AddLessonBlockInput, error) {
	lessonID, err := requiredFlag(command, "lesson-id")
	if err != nil {
		return core.AddLessonBlockInput{}, err
	}
	kind, err := requiredFlag(command, "kind")
	if err != nil {
		return core.AddLessonBlockInput{}, err
	}

	input := core.AddLessonBlockInput{
		LessonID: lessonID,
		Kind:     kind,
		Position: changedIntFlag(command, "position"),
	}

	switch strings.TrimSpace(kind) {
	case "text":
		markdown, err := requiredChangedStringFlag(command, "text")
		if err != nil {
			return core.AddLessonBlockInput{}, err
		}
		input.Markdown = markdown
	case "video":
		provider, err := requiredFlag(command, "video-provider")
		if err != nil {
			return core.AddLessonBlockInput{}, err
		}
		locator, err := requiredFlag(command, "video-locator")
		if err != nil {
			return core.AddLessonBlockInput{}, err
		}
		input.VideoProvider = provider
		input.VideoLocator = locator
		input.VideoCaption = stringFlag(command, "video-caption")
	case "quiz":
		quizID, err := requiredFlag(command, "quiz-id")
		if err != nil {
			return core.AddLessonBlockInput{}, err
		}
		input.QuizRef = quizID
	case "practice":
		practiceID, err := requiredFlag(command, "practice-id")
		if err != nil {
			return core.AddLessonBlockInput{}, err
		}
		input.PracticeRef = practiceID
	}

	return input, nil
}

func newLessonBlockListCommand(context lessonCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "list",
		Short: "List lesson blocks",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			lessonID, err := requiredFlag(command, "lesson-id")
			if err != nil {
				return err
			}

			out, err := context.service.ListLessonBlocks(core.ListLessonBlocksInput{LessonID: lessonID})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderBlockList(outputFormat(command), out.Blocks)
		},
	}

	flags := command.Flags()
	flags.String("lesson-id", "", "lesson id")
	flags.StringP("output", "o", "table", "output format")

	return command
}

func newLessonBlockGetCommand(context lessonCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "get <block-id>",
		Short: "Get a lesson block",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			out, err := context.service.GetLessonBlock(core.GetLessonBlockInput{ID: args[0]})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderBlock(outputFormat(command), out.Block)
		},
	}

	command.Flags().StringP("output", "o", "table", "output format")

	return command
}

func newLessonBlockUpdateCommand(context lessonCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "update <block-id>",
		Short: "Update a lesson block",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			out, err := context.service.UpdateLessonBlock(core.UpdateLessonBlockInput{
				ID:            args[0],
				Markdown:      changedStringFlag(command, "text"),
				VideoProvider: changedStringFlag(command, "video-provider"),
				VideoLocator:  changedStringFlag(command, "video-locator"),
				VideoCaption:  changedStringFlag(command, "video-caption"),
			})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderUpdatedBlock(out.ID)
		},
	}

	flags := command.Flags()
	flags.String("text", "", "text markdown")
	flags.String("video-provider", "", "video provider")
	flags.String("video-locator", "", "video locator")
	flags.String("video-caption", "", "video caption")

	return command
}

func newLessonBlockRemoveCommand(context lessonCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "remove <block-id>",
		Short: "Remove a lesson block",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			if !boolFlag(command, "force") {
				confirmed, err := context.prompter.Confirm(fmt.Sprintf("remove lesson block %s", args[0]))
				if err != nil {
					return err
				}
				if !confirmed {
					return ErrConfirmationDeclined
				}
			}

			if err := context.service.RemoveLessonBlock(core.RemoveLessonBlockInput{ID: args[0]}); err != nil {
				return err
			}

			return context.rendererFor(command).RenderConfirmation("lesson block removed")
		},
	}

	command.Flags().Bool("force", false, "skip confirmation")

	return command
}

func newLessonBlockReorderCommand(context lessonCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "reorder",
		Short: "Reorder lesson blocks",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			lessonID, err := requiredFlag(command, "lesson-id")
			if err != nil {
				return err
			}

			order, err := requiredFlag(command, "order")
			if err != nil {
				return err
			}
			placements, err := parseBlockPlacements(order)
			if err != nil {
				return err
			}

			if err := context.service.ReorderLessonBlocks(core.ReorderLessonBlocksInput{
				LessonID: lessonID,
				Order:    placements,
			}); err != nil {
				return err
			}

			return context.rendererFor(command).RenderConfirmation("lesson blocks reordered")
		},
	}

	flags := command.Flags()
	flags.String("lesson-id", "", "lesson id")
	flags.String("order", "", "block-id:position pairs")

	return command
}

func parseBlockPlacements(value string) ([]core.BlockPlacementDTO, error) {
	parts := strings.Split(value, ",")
	placements := make([]core.BlockPlacementDTO, 0, len(parts))

	for _, part := range parts {
		placement, err := parseBlockPlacement(part)
		if err != nil {
			return nil, err
		}

		placements = append(placements, placement)
	}

	return placements, nil
}

func parseBlockPlacement(value string) (core.BlockPlacementDTO, error) {
	blockID, positionText, ok := strings.Cut(value, ":")
	if !ok {
		return core.BlockPlacementDTO{}, fmt.Errorf("%w: %s", ErrInvalidBlockOrder, value)
	}

	blockID = strings.TrimSpace(blockID)
	positionText = strings.TrimSpace(positionText)
	if blockID == "" || positionText == "" {
		return core.BlockPlacementDTO{}, fmt.Errorf("%w: %s", ErrInvalidBlockOrder, value)
	}

	position, err := strconv.Atoi(positionText)
	if err != nil {
		return core.BlockPlacementDTO{}, fmt.Errorf("%w: %s", ErrInvalidBlockOrder, value)
	}

	return core.BlockPlacementDTO{BlockID: blockID, Position: position}, nil
}
