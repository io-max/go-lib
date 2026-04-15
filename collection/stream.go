package collection

type Stream[T any] struct {
	executor func() []T
	data     []T
}

func Of[T any](collection []T) *Stream[T] {
	if collection == nil {
		collection = []T{}
	}
	return &Stream[T]{
		data: collection,
		executor: func() []T {
			return collection
		},
	}
}

func (s *Stream[T]) Collect() []T {
	if s.executor != nil {
		s.data = s.executor()
		s.executor = nil
	}
	result := make([]T, len(s.data))
	copy(result, s.data)
	return result
}

func Map[T, U any](s *Stream[T], fn func(T) U) *Stream[U] {
	executor := func() []U {
		data := s.Collect()
		result := make([]U, len(data))
		for i, item := range data {
			result[i] = fn(item)
		}
		return result
	}
	return &Stream[U]{
		executor: executor,
	}
}
