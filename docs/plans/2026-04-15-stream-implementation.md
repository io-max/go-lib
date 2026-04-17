# Stream API Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement a complete generic Stream API for Go with lazy evaluation, method chaining, and comprehensive collection operations.

**Architecture:** Modify existing `collection/stream.go` with lazy evaluation via function composition. Each transformation returns a new Stream wrapping a closure. Short-circuit operations terminate early. Use defensive nil/empty handling throughout.

**Tech Stack:** Go 1.x, standard library only, testing with `testing` package.

---

## Task 1: Restructure Stream for Lazy Evaluation

**Files:**
- Modify: `collection/stream.go`

**Step 1: Read current implementation**

```go
type Stream[T any] struct {
    data []T
}
```

**Step 2: Restructure to lazy evaluation**

```go
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
```

**Step 3: Run tests to verify nothing breaks**

Run: `go test ./collection/...`
Expected: PASS

**Step 4: Commit**

```bash
git add collection/stream.go
git commit -m "refactor(collection): restructure Stream for lazy evaluation"
```

---

## Task 2: Implement Filter

**Files:**
- Modify: `collection/stream.go`
- Modify: `collection/stream_test.go`

**Step 1: Write failing test**

```go
func TestFilter(t *testing.T) {
    nums := []int{1, 2, 3, 4, 5}
    evens := Of(nums).Filter(func(n int) bool { return n%2 == 0 }).Collect()
    if len(evens) != 2 || evens[0] != 2 || evens[1] != 4 {
        t.Errorf("expected [2 4], got %v", evens)
    }
}

func TestFilterWithNil(t *testing.T) {
    evens := Of([]int(nil)).Filter(func(n int) bool { return n%2 == 0 }).Collect()
    if len(evens) != 0 {
        t.Errorf("expected [], got %v", evens)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./collection/... -run TestFilter -v`
Expected: FAIL with "Filter not defined"

**Step 3: Implement Filter method**

```go
func (s *Stream[T]) Filter(pred func(T) bool) *Stream[T] {
    return &Stream[T]{
        data: s.data,
        executor: func() []T {
            result := []T{}
            for _, item := range s.executor() {
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
```

**Step 4: Run tests to verify they pass**

Run: `go test ./collection/... -run TestFilter -v`
Expected: PASS

**Step 5: Commit**

```bash
git add collection/stream.go collection/stream_test.go
git commit -m "feat(collection): add Filter method"
```

---

## Task 3: Implement FlatMap

**Files:**
- Modify: `collection/stream.go`
- Modify: `collection/stream_test.go`

**Step 1: Write failing test**

```go
func TestFlatMap(t *testing.T) {
    nested := [][]int{{1, 2}, {3, 4}, {5}}
    flat := Of(nested).FlatMap(func(s []int) []int { return s }).Collect()
    if len(flat) != 5 || flat[0] != 1 || flat[4] != 5 {
        t.Errorf("expected [1 2 3 4 5], got %v", flat)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./collection/... -run TestFlatMap -v`
Expected: FAIL with "FlatMap not defined"

**Step 3: Implement FlatMap**

```go
func (s *Stream[T]) FlatMap(fn func(T) []U) *Stream[U] {
    return &Stream[U]{
        data: nil,
        executor: func() []U {
            result := []U{}
            for _, item := range s.executor() {
                result = append(result, fn(item)...)
            }
            return result
        },
    }
}

func FlatMap[T, U any](s *Stream[T], fn func(T) []U) *Stream[U] {
    return s.FlatMap(fn)
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./collection/... -run TestFlatMap -v`
Expected: PASS

**Step 5: Commit**

```bash
git add collection/stream.go collection/stream_test.go
git commit -m "feat(collection): add FlatMap method"
```

---

## Task 4: Implement Distinct

**Files:**
- Modify: `collection/stream.go`
- Modify: `collection/stream_test.go`

**Step 1: Write failing test**

```go
func TestDistinct(t *testing.T) {
    nums := []int{1, 2, 2, 3, 3, 3}
    distinct := Of(nums).Distinct().Collect()
    if len(distinct) != 3 {
        t.Errorf("expected 3 distinct elements, got %v", distinct)
    }
}

func TestDistinctWithComparer(t *testing.T) {
    users := []User{{ID: 1, Name: "a"}, {ID: 1, Name: "b"}, {ID: 2, Name: "c"}}
    distinct := Distinct(Of(users), func(a, b User) bool { return a.ID == b.ID }).Collect()
    if len(distinct) != 2 {
        t.Errorf("expected 2 distinct users by ID, got %v", distinct)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./collection/... -run TestDistinct -v`
Expected: FAIL with "Distinct not defined"

**Step 3: Implement Distinct**

```go
func (s *Stream[T]) Distinct() *Stream[T] {
    return &Stream[T]{
        data: s.data,
        executor: func() []T {
            seen := make(map[string]bool)
            result := []T{}
            for _, item := range s.executor() {
                key := fmt.Sprintf("%v", item)
                if !seen[key] {
                    seen[key] = true
                    result = append(result, item)
                }
            }
            return result
        },
    }
}

func Distinct[T any](s *Stream[T], eq func(T, T) bool) *Stream[T] {
    return &Stream[T]{
        data: s.data,
        executor: func() []T {
            result := []T{}
            for _, item := range s.executor() {
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
```

**Step 4: Run tests to verify they pass**

Run: `go test ./collection/... -run TestDistinct -v`
Expected: PASS

**Step 5: Commit**

```bash
git add collection/stream.go collection/stream_test.go
git commit -m "feat(collection): add Distinct method"
```

---

## Task 5: Implement Peek

**Files:**
- Modify: `collection/stream.go`
- Modify: `collection/stream_test.go`

**Step 1: Write failing test**

```go
func TestPeek(t *testing.T) {
    var peeked []int
    nums := []int{1, 2, 3}
    Of(nums).Peek(func(n int) { peeked = append(peeked, n) }).Collect()
    if len(peeked) != 3 || peeked[0] != 1 || peeked[2] != 3 {
        t.Errorf("expected [1 2 3], got %v", peeked)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./collection/... -run TestPeek -v`
Expected: FAIL with "Peek not defined"

**Step 3: Implement Peek**

```go
func (s *Stream[T]) Peek(fn func(T)) *Stream[T] {
    return &Stream[T]{
        data: s.data,
        executor: func() []T {
            result := s.executor()
            for _, item := range result {
                fn(item)
            }
            return result
        },
    }
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./collection/... -run TestPeek -v`
Expected: PASS

**Step 5: Commit**

```bash
git add collection/stream.go collection/stream_test.go
git commit -m "feat(collection): add Peek method"
```

---

## Task 6: Implement Aggregation Operations

**Files:**
- Modify: `collection/stream.go`
- Modify: `collection/stream_test.go`

**Step 1: Write failing tests**

```go
func TestCount(t *testing.T) {
    count := Of([]int{1, 2, 3, 4, 5}).Count()
    if count != 5 {
        t.Errorf("expected 5, got %d", count)
    }
}

func TestReduce(t *testing.T) {
    sum := Of([]int{1, 2, 3, 4}).Reduce(0, func(acc, n int) int { return acc + n })
    if sum != 10 {
        t.Errorf("expected 10, got %d", sum)
    }
}

func TestFindFirst(t *testing.T) {
    first, ok := Of([]int{1, 2, 3}).Filter(func(n int) bool { return n%2 == 0 }).FindFirst()
    if !ok || first != 2 {
        t.Errorf("expected (2, true), got (%d, %v)", first, ok)
    }
    
    _, ok = Of([]int{1, 3, 5}).Filter(func(n int) bool { return n%2 == 0 }).FindFirst()
    if ok {
        t.Errorf("expected (0, false), got (%v)", ok)
    }
}

func TestAnyMatch(t *testing.T) {
    if !Of([]int{1, 2, 3}).AnyMatch(func(n int) bool { return n == 2 }) {
        t.Errorf("expected true")
    }
    if Of([]int{1, 2, 3}).AnyMatch(func(n int) bool { return n == 5 }) {
        t.Errorf("expected false")
    }
}

func TestAllMatch(t *testing.T) {
    if !Of([]int{1, 2, 3}).AllMatch(func(n int) bool { return n > 0 }) {
        t.Errorf("expected true")
    }
    if Of([]int{1, 2, 3}).AllMatch(func(n int) bool { return n > 1 }) {
        t.Errorf("expected false")
    }
}

func TestNoneMatch(t *testing.T) {
    if !Of([]int{1, 2, 3}).NoneMatch(func(n int) bool { return n > 10 }) {
        t.Errorf("expected true")
    }
    if Of([]int{1, 2, 3}).NoneMatch(func(n int) bool { return n == 2 }) {
        t.Errorf("expected false")
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./collection/... -run "TestCount|TestReduce|TestFindFirst|TestAnyMatch|TestAllMatch|TestNoneMatch" -v`
Expected: FAIL with "not defined"

**Step 3: Implement aggregation methods**

```go
func (s *Stream[T]) Count() int {
    return len(s.executor())
}

func (s *Stream[T]) Reduce(initial T, fn func(T, T) T) T {
    result := initial
    for _, item := range s.executor() {
        result = fn(result, item)
    }
    return result
}

func (s *Stream[T]) FindFirst() (T, bool) {
    data := s.executor()
    if len(data) == 0 {
        return *new(T), false
    }
    return data[0], true
}

func (s *Stream[T]) AnyMatch(pred func(T) bool) bool {
    for _, item := range s.executor() {
        if pred(item) {
            return true
        }
    }
    return false
}

func (s *Stream[T]) AllMatch(pred func(T) bool) bool {
    for _, item := range s.executor() {
        if !pred(item) {
            return false
        }
    }
    return true
}

func (s *Stream[T]) NoneMatch(pred func(T) bool) bool {
    return !s.AnyMatch(pred)
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./collection/... -run "TestCount|TestReduce|TestFindFirst|TestAnyMatch|TestAllMatch|TestNoneMatch" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add collection/stream.go collection/stream_test.go
git commit -m "feat(collection): add aggregation operations (Count, Reduce, FindFirst, AnyMatch, AllMatch, NoneMatch)"
```

---

## Task 7: Implement Data Organization Operations

**Files:**
- Modify: `collection/stream.go`
- Modify: `collection/stream_test.go`

**Step 1: Write failing tests**

```go
func TestLimit(t *testing.T) {
    limited := Of([]int{1, 2, 3, 4, 5}).Limit(3).Collect()
    if len(limited) != 3 {
        t.Errorf("expected 3 elements, got %v", limited)
    }
}

func TestLimitWithZero(t *testing.T) {
    limited := Of([]int{1, 2, 3}).Limit(0).Collect()
    if len(limited) != 0 {
        t.Errorf("expected empty, got %v", limited)
    }
}

func TestSkip(t *testing.T) {
    skipped := Of([]int{1, 2, 3, 4, 5}).Skip(2).Collect()
    if len(skipped) != 3 {
        t.Errorf("expected 3 elements, got %v", skipped)
    }
}

func TestSkipMoreThanLength(t *testing.T) {
    skipped := Of([]int{1, 2, 3}).Skip(10).Collect()
    if len(skipped) != 0 {
        t.Errorf("expected empty, got %v", skipped)
    }
}

func TestSorted(t *testing.T) {
    sorted := Of([]int{3, 1, 2}).Sorted(func(a, b int) bool { return a < b }).Collect()
    if sorted[0] != 1 || sorted[1] != 2 || sorted[2] != 3 {
        t.Errorf("expected [1 2 3], got %v", sorted)
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./collection/... -run "TestLimit|TestSkip|TestSorted" -v`
Expected: FAIL with "not defined"

**Step 3: Implement data organization methods**

```go
func (s *Stream[T]) Limit(n int) *Stream[T] {
    if n <= 0 {
        return Of([]T{})
    }
    return &Stream[T]{
        data: s.data,
        executor: func() []T {
            result := []T{}
            for i, item := range s.executor() {
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
        data: s.data,
        executor: func() []T {
            result := []T{}
            for i, item := range s.executor() {
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
            data := s.executor()
            sorted := make([]T, len(data))
            copy(sorted, data)
            sort.Slice(sorted, func(i, j int) bool {
                return less(sorted[i], sorted[j])
            })
            return sorted
        },
    }
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./collection/... -run "TestLimit|TestSkip|TestSorted" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add collection/stream.go collection/stream_test.go
git commit -m "feat(collection): add data organization operations (Limit, Skip, Sorted)"
```

---

## Task 8: Implement Traversal Operations

**Files:**
- Modify: `collection/stream.go`
- Modify: `collection/stream_test.go`

**Step 1: Write failing tests**

```go
func TestForEach(t *testing.T) {
    var sum int
    Of([]int{1, 2, 3}).ForEach(func(n int) { sum += n })
    if sum != 6 {
        t.Errorf("expected 6, got %d", sum)
    }
}

func TestToMap(t *testing.T) {
    users := []User{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}}
    m := Of(users).ToMap(func(u User) int64 { return u.ID })
    if len(m) != 2 || m[1].Name != "a" {
        t.Errorf("expected map with 2 entries, got %v", m)
    }
}

func TestToMapWithValue(t *testing.T) {
    users := []User{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}}
    m := Of(users).ToMapWithValue(func(u User) int64 { return u.ID }, func(u User) string { return u.Name })
    if len(m) != 2 || m[1] != "a" || m[2] != "b" {
        t.Errorf("expected map with 2 entries, got %v", m)
    }
}

func TestToSet(t *testing.T) {
    nums := []int{1, 2, 2, 3, 3, 3}
    s := Of(nums).ToSet()
    if len(s) != 3 {
        t.Errorf("expected set with 3 entries, got %v", s)
    }
    if _, ok := s[1]; !ok {
        t.Errorf("expected 1 in set")
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./collection/... -run "TestForEach|TestToMap|TestToSet" -v`
Expected: FAIL with "not defined"

**Step 3: Implement traversal methods**

```go
func (s *Stream[T]) ForEach(fn func(T)) {
    for _, item := range s.executor() {
        fn(item)
    }
}

func (s *Stream[T]) ToMap(keyFn func(T) K) map[K]T {
    result := make(map[K]T)
    for _, item := range s.executor() {
        result[keyFn(item)] = item
    }
    return result
}

func (s *Stream[T]) ToMapWithValue(keyFn func(T) K, valueFn func(T) V) map[K]V {
    result := make(map[K]V)
    for _, item := range s.executor() {
        result[keyFn(item)] = valueFn(item)
    }
    return result
}

func (s *Stream[T]) ToSet() map[T]struct{} {
    result := make(map[T]struct{})
    for _, item := range s.executor() {
        result[item] = struct{}{}
    }
    return result
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./collection/... -run "TestForEach|TestToMap|TestToSet" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add collection/stream.go collection/stream_test.go
git commit -m "feat(collection): add traversal operations (ForEach, ToMap, ToSet)"
```

---

## Task 9: Add import for fmt and sort

**Files:**
- Modify: `collection/stream.go`

**Step 1: Add required imports**

Ensure `fmt` and `sort` packages are imported for Distinct and Sorted implementations.

**Step 2: Run all tests**

Run: `go test ./collection/... -v`
Expected: PASS

**Step 3: Commit**

```bash
git add collection/stream.go
git commit -m "chore(collection): add fmt and sort imports"
```

---

## Task 10: Final Verification

**Step 1: Run all tests**

Run: `go test ./collection/... -v`
Expected: All tests pass

**Step 2: Run go vet**

Run: `go vet ./collection/...`
Expected: No issues

**Step 3: Run golint if available**

Run: `golangci-lint run ./collection/...` (or skip if not installed)

**Step 4: Final commit**

```bash
git add -A
git commit -m "feat(collection): complete Stream API implementation"
```
