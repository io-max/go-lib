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

func (s *Stream[T]) Filter(pred func(T) bool) *Stream[T] {
	return &Stream[T]{
		data: nil,
		executor: func() []T {
			result := []T{}
			for _, item := range s.Collect() {
				if pred(item) {
					result = append(result, item)
				}
			}
			return result
		},
	}
}

func Filter[T any](s *Stream[T], pred func(T) bool) *Stream[T] {
	return s.Filter(pred)
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
