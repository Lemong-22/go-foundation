package domain

type LessonOrder struct {
	value int
}

func NewLessonOrder(value int) (LessonOrder, error) {
	if value < 0 {
		return LessonOrder{}, NewValidationError("order", "must be greater than or equal to zero")
	}

	return LessonOrder{value: value}, nil
}

func (order LessonOrder) Int() int {
	return order.value
}
