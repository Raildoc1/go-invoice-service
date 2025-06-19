package flagtypes

type Argument[T any] struct {
	val *T
}

func newArgument[T any]() *Argument[T] {
	return &Argument[T]{
		val: nil,
	}
}

func (a *Argument[T]) Value() (T, bool) {
	if a.val == nil {
		var zero T
		return zero, false
	}
	return *a.val, true
}
