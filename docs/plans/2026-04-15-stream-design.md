# Stream API Design

## Overview

A generic Stream API library for Go, providing functional-style collection operations. Designed as a general-purpose utility library suitable for various business scenarios.

## Design Decisions

### API Style: Combinable

- Core methods return `*Stream[T]` supporting method chaining
- Standalone function versions `Filter`, `Map` etc. also available
- Flexible usage: both chaining and function composition supported

```go
// Method chaining
Of(users).Filter(func(u User) bool { return u.ID > 0 }).Map(ToName).Collect()

// Function composition
Filter(Map(Of(users), ToName), func(id int64) bool { return id > 0 }).Collect()
```

### Evaluation Strategy: Lazy + Short-circuit

- Operations are lazily evaluated via function composition
- Short-circuit operations (AnyMatch, FindFirst, etc.) stop traversal early

### Error Handling: Defensive

| Scenario | Behavior |
|----------|----------|
| nil input | Converted to empty slice |
| Empty Stream | Returns safe empty result |
| Limit(n) where n <= 0 | Returns empty Stream |
| Skip(n) where n >= len | Returns empty Stream |
| FindFirst() on empty | Returns (zeroValue, false) |

- No panic on boundary conditions
- Callers responsible for checking return values

## Operations

### Creation
- `Of[T any](collection []T) *Stream[T]` - Create stream from slice
- `(s *Stream[T]) Collect() []T` - Materialize stream result

### Transformation
- `(s *Stream[T]) Filter(pred func(T) bool) *Stream[T]`
- `(s *Stream[T]) Map(fn func(T) U) *Stream[U]`
- `(s *Stream[T]) FlatMap(fn func(T) []U) *Stream[U]`
- `(s *Stream[T]) Distinct() *Stream[T]`
- `(s *Stream[T]) Peek(fn func(T)) *Stream[T]`

### Aggregation
- `(s *Stream[T]) Count() int`
- `(s *Stream[T]) Reduce(initial T, fn func(T, T) T) T`
- `(s *Stream[T]) FindFirst() (T, bool)`
- `(s *Stream[T]) AnyMatch(pred func(T) bool) bool`
- `(s *Stream[T]) AllMatch(pred func(T) bool) bool`
- `(s *Stream[T]) NoneMatch(pred func(T) bool) bool`

### Data Organization
- `(s *Stream[T]) Limit(n int) *Stream[T]`
- `(s *Stream[T]) Skip(n int) *Stream[T]`
- `(s *Stream[T]) Sorted(less func(T, T) bool) *Stream[T]`

### Traversal
- `(s *Stream[T]) ForEach(fn func(T))`
- `(s *Stream[T]) ToMap(keyFn func(T) K) map[K]T`
- `(s *Stream[T]) ToMapWithValue(keyFn func(T) K, valueFn func(T) V) map[K]V`
- `(s *Stream[T]) ToSet() map[T]struct{}`

## Implementation Notes

- Uses function composition for lazy evaluation
- Short-circuit operations traverse until condition is met
- Sorted triggers eager evaluation (requires full data for sorting)
