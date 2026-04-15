package collection

import (
	"reflect"
	"sort"
)

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

func FlatMap[T, U any](s *Stream[T], fn func(T) []U) *Stream[U] {
	return &Stream[U]{
		data: nil,
		executor: func() []U {
			result := []U{}
			for _, item := range s.Collect() {
				result = append(result, fn(item)...)
			}
			return result
		},
	}
}

func (s *Stream[T]) Distinct() *Stream[T] {
	return &Stream[T]{
		data: nil,
		executor: func() []T {
			result := []T{}
			for _, item := range s.Collect() {
				found := false
				for _, existing := range result {
					if reflect.DeepEqual(item, existing) {
						found = true
						break
					}
				}
				if !found {
					result = append(result, item)
				}
			}
			return result
		},
	}
}

func Distinct[T any](s *Stream[T], eq func(T, T) bool) *Stream[T] {
	return &Stream[T]{
		data: nil,
		executor: func() []T {
			result := []T{}
			for _, item := range s.Collect() {
				found := false
				for _, existing := range result {
					if eq(item, existing) {
						found = true
						break
					}
				}
				if !found {
					result = append(result, item)
				}
			}
			return result
		},
	}
}

func (s *Stream[T]) Peek(fn func(T)) *Stream[T] {
	return &Stream[T]{
		data: nil,
		executor: func() []T {
			result := s.Collect()
			for _, item := range result {
				fn(item)
			}
			return result
		},
	}
}

func (s *Stream[T]) Count() int {
	return len(s.Collect())
}

func (s *Stream[T]) Reduce(initial T, fn func(T, T) T) T {
	result := initial
	for _, item := range s.Collect() {
		result = fn(result, item)
	}
	return result
}

func (s *Stream[T]) FindFirst() (T, bool) {
	data := s.Collect()
	if len(data) == 0 {
		return *new(T), false
	}
	return data[0], true
}

func (s *Stream[T]) AnyMatch(pred func(T) bool) bool {
	for _, item := range s.Collect() {
		if pred(item) {
			return true
		}
	}
	return false
}

func (s *Stream[T]) AllMatch(pred func(T) bool) bool {
	for _, item := range s.Collect() {
		if !pred(item) {
			return false
		}
	}
	return true
}

func (s *Stream[T]) NoneMatch(pred func(T) bool) bool {
	return !s.AnyMatch(pred)
}

func (s *Stream[T]) Limit(n int) *Stream[T] {
	if n <= 0 {
		return Of([]T{})
	}
	return &Stream[T]{
		data: nil,
		executor: func() []T {
			result := []T{}
			for i, item := range s.Collect() {
				if i >= n {
					break
				}
				result = append(result, item)
			}
			return result
		},
	}
}

func (s *Stream[T]) Skip(n int) *Stream[T] {
	if n <= 0 {
		return s
	}
	return &Stream[T]{
		data: nil,
		executor: func() []T {
			result := []T{}
			for i, item := range s.Collect() {
				if i >= n {
					result = append(result, item)
				}
			}
			return result
		},
	}
}

func (s *Stream[T]) Sorted(less func(T, T) bool) *Stream[T] {
	return &Stream[T]{
		data: nil,
		executor: func() []T {
			data := s.Collect()
			sorted := make([]T, len(data))
			copy(sorted, data)
			sort.Slice(sorted, func(i, j int) bool {
				return less(sorted[i], sorted[j])
			})
			return sorted
		},
	}
}

func (s *Stream[T]) ForEach(fn func(T)) {
	for _, item := range s.Collect() {
		fn(item)
	}
}

func ToMap[T any, K comparable](s *Stream[T], keyFn func(T) K) map[K]T {
	result := make(map[K]T)
	for _, item := range s.Collect() {
		result[keyFn(item)] = item
	}
	return result
}

func ToMapWithValue[T any, K comparable, V any](s *Stream[T], keyFn func(T) K, valueFn func(T) V) map[K]V {
	result := make(map[K]V)
	for _, item := range s.Collect() {
		result[keyFn(item)] = valueFn(item)
	}
	return result
}

func ToSet[T comparable](s *Stream[T]) map[T]struct{} {
	result := make(map[T]struct{})
	for _, item := range s.Collect() {
		result[item] = struct{}{}
	}
	return result
}
