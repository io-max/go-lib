package collection

type Stream[T any] struct {
	data []T
}

func Of[T any](collection []T) *Stream[T] {
	if collection == nil {
		collection = []T{}
	}
	return &Stream[T]{data: collection}
}

func (s *Stream[T]) Collect() []T {
	result := make([]T, len(s.data))
	copy(result, s.data)
	return result
}

func Map[T, U any](s *Stream[T], fn func(T) U) *Stream[U] {
	result := make([]U, len(s.data))
	for i, item := range s.data {
		result[i] = fn(item)
	}
	return &Stream[U]{data: result}
}
