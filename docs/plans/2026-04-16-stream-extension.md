# Stream API Extension Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extend Stream API with type conversions, set operations, and GroupBy functionality.

**Architecture:** Add new standalone functions to `collection/stream.go`. Use `E` suffix for error-returning variants. Follow existing lazy evaluation pattern where applicable.

**Tech Stack:** Go 1.x, standard library (strconv, fmt, errors)

---

## Task 1: Numeric to String Conversions

**Files:**
- Modify: `collection/stream.go`
- Modify: `collection/stream_test.go`

**Step 1: Write failing tests**

```go
func TestToString(t *testing.T) {
    result := ToString(Of([]int64{1, 2, 3})).Collect()
    if len(result) != 3 || result[0] != "1" || result[1] != "2" || result[2] != "3" {
        t.Errorf("expected [1 2 3], got %v", result)
    }
}

func TestToStringE(t *testing.T) {
    result, err := ToStringE(Of([]int64{1, 2, 3}))
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
    if result.Collect()[0] != "1" {
        t.Errorf("expected [1 2 3]")
    }
}
```

**Step 2: Run tests**

Run: `go test ./collection/... -run TestToString -v`
Expected: FAIL

**Step 3: Implement**

```go
func ToString[T ~int | ~int64 | ~int32 | ~float64 | ~float32](s *Stream[T]) *Stream[string] {
    return Map(s, func(v T) string { return fmt.Sprintf("%v", v) })
}

func ToStringE[T ~int | ~int64 | ~int32 | ~float64 | ~float32](s *Stream[T]) (*Stream[string], error) {
    return ToString(s), nil
}
```

**Step 4: Run tests**
Expected: PASS

**Step 5: Commit**

```bash
git add collection/stream.go collection/stream_test.go
git commit -m "feat(collection): add ToString conversion"
```

---

## Task 2: String to Numeric Conversions

**Files:**
- Modify: `collection/stream.go`
- Modify: `collection/stream_test.go`

**Step 1: Write failing tests**

```go
func TestStringToInt64(t *testing.T) {
    result := StringToInt64(Of([]string{"1", "2", "3"})).Collect()
    if len(result) != 3 || result[0] != 1 || result[1] != 2 || result[2] != 3 {
        t.Errorf("expected [1 2 3], got %v", result)
    }
}

func TestStringToInt64WithError(t *testing.T) {
    _, err := StringToInt64E(Of([]string{"1", "a", "3"}))
    if err == nil {
        t.Errorf("expected error for invalid conversion")
    }
}
```

**Step 2: Run tests**
Expected: FAIL

**Step 3: Implement**

```go
func StringToInt64(s *Stream[string]) *Stream[int64] {
    return Map(s, func(v string) int64 {
        n, _ := strconv.ParseInt(v, 10, 64)
        return n
    })
}

func StringToInt64E(s *Stream[string]) (*Stream[int64], error) {
    var err error
    result := Map(s, func(v string) int64 {
        n, e := strconv.ParseInt(v, 10, 64)
        if e != nil {
            err = e
        }
        return n
    }).Collect()
    if err != nil {
        return nil, err
    }
    return Of(result), nil
}

func StringToFloat64(s *Stream[string]) *Stream[float64] {
    return Map(s, func(v string) float64 {
        f, _ := strconv.ParseFloat(v, 64)
        return f
    })
}

func StringToFloat64E(s *Stream[string]) (*Stream[float64], error) {
    var err error
    result := Map(s, func(v string) float64 {
        f, e := strconv.ParseFloat(v, 64)
        if e != nil {
            err = e
        }
        return f
    }).Collect()
    if err != nil {
        return nil, err
    }
    return Of(result), nil
}
```

**Step 4: Run tests**
Expected: PASS

**Step 5: Commit**

```bash
git add collection/stream.go collection/stream_test.go
git commit -m "feat(collection): add StringToInt64 and StringToFloat64 conversions"
```

---

## Task 3: Numeric Type Conversions

**Files:**
- Modify: `collection/stream.go`
- Modify: `collection/stream_test.go`

**Step 1: Write failing tests**

```go
func TestToInt(t *testing.T) {
    result := ToInt(Of([]int64{1, 2, 3})).Collect()
    if len(result) != 3 || result[0] != 1 {
        t.Errorf("expected [1 2 3], got %v", result)
    }
}

func TestToFloat64(t *testing.T) {
    result := ToFloat64(Of([]int64{1, 2, 3})).Collect()
    if len(result) != 3 || result[0] != 1.0 {
        t.Errorf("expected [1.0 2.0 3.0], got %v", result)
    }
}
```

**Step 2: Run tests**
Expected: FAIL

**Step 3: Implement**

```go
func ToInt[T ~int64 | ~int32](s *Stream[T]) *Stream[int] {
    return Map(s, func(v T) int { return int(v) })
}

func ToIntE[T ~int64 | ~int32](s *Stream[T]) (*Stream[int], error) {
    return ToInt(s), nil
}

func ToFloat64[T ~int | ~int64 | ~int32 | ~float64 | ~float32](s *Stream[T]) *Stream[float64] {
    return Map(s, func(v T) float64 { return float64(v) })
}

func ToFloat64E[T ~int | ~int64 | ~int32 | ~float64 | ~float32](s *Stream[T]) (*Stream[float64], error) {
    return ToFloat64(s), nil
}
```

**Step 4: Run tests**
Expected: PASS

**Step 5: Commit**

```bash
git add collection/stream.go collection/stream_test.go
git commit -m "feat(collection): add ToInt and ToFloat64 conversions"
```

---

## Task 4: HexString and Bytes Conversions

**Files:**
- Modify: `collection/stream.go`
- Modify: `collection/stream_test.go`

**Step 1: Write failing tests**

```go
func TestToHexString(t *testing.T) {
    result := ToHexString(Of([]int64{255, 16})).Collect()
    if len(result) != 2 || result[0] != "ff" || result[1] != "10" {
        t.Errorf("expected [ff 10], got %v", result)
    }
}

func TestToBytes(t *testing.T) {
    result := ToBytes(Of([]string{"hello", "world"})).Collect()
    expected := [][]byte{[]byte("hello"), []byte("world")}
    if len(result) != 2 || string(result[0]) != "hello" {
        t.Errorf("expected [hello world], got %v", result)
    }
}
```

**Step 2: Run tests**
Expected: FAIL

**Step 3: Implement**

```go
func ToHexString[T ~int | ~int64 | ~int32](s *Stream[T]) *Stream[string] {
    return Map(s, func(v T) string { return fmt.Sprintf("%x", v) })
}

func ToHexStringE[T ~int | ~int64 | ~int32](s *Stream[T]) (*Stream[string], error) {
    return ToHexString(s), nil
}

func ToBytes(s *Stream[string]) *Stream[[]byte] {
    return Map(s, func(v string) []byte { return []byte(v) })
}

func ToBytesE(s *Stream[string]) (*Stream[[]byte], error) {
    return ToBytes(s), nil
}
```

**Step 4: Run tests**
Expected: PASS

**Step 5: Commit**

```bash
git add collection/stream.go collection/stream_test.go
git commit -m "feat(collection): add ToHexString and ToBytes conversions"
```

---

## Task 5: Generic Convert Function

**Files:**
- Modify: `collection/stream.go`
- Modify: `collection/stream_test.go`

**Step 1: Write failing tests**

```go
func TestConvert(t *testing.T) {
    result := Convert(Of([]int{1, 2, 3}), func(i int) string {
        return fmt.Sprintf("num_%d", i)
    }).Collect()
    expected := []string{"num_1", "num_2", "num_3"}
    if len(result) != 3 || result[0] != "num_1" {
        t.Errorf("expected %v, got %v", expected, result)
    }
}
```

**Step 2: Run tests**
Expected: FAIL

**Step 3: Implement**

```go
func Convert[T any, U any](s *Stream[T], fn func(T) U) *Stream[U] {
    return Map(s, fn)
}
```

**Step 4: Run tests**
Expected: PASS

**Step 5: Commit**

```bash
git add collection/stream.go collection/stream_test.go
git commit -m "feat(collection): add Convert generic function"
```

---

## Task 6: Set Operations - Union, Intersect, Difference

**Files:**
- Modify: `collection/stream.go`
- Modify: `collection/stream_test.go`

**Step 1: Write failing tests**

```go
func TestUnion(t *testing.T) {
    a := Of([]int{1, 2, 3})
    b := Of([]int{2, 3, 4})
    result := Union(a, b).Collect()
    if len(result) != 4 {
        t.Errorf("expected 4 elements, got %v", result)
    }
}

func TestIntersect(t *testing.T) {
    a := Of([]int{1, 2, 3})
    b := Of([]int{2, 3, 4})
    result := Intersect(a, b).Collect()
    if len(result) != 2 || result[0] != 2 || result[1] != 3 {
        t.Errorf("expected [2 3], got %v", result)
    }
}

func TestDifference(t *testing.T) {
    a := Of([]int{1, 2, 3})
    b := Of([]int{2, 3, 4})
    result := Difference(a, b).Collect()
    if len(result) != 1 || result[0] != 1 {
        t.Errorf("expected [1], got %v", result)
    }
}
```

**Step 2: Run tests**
Expected: FAIL

**Step 3: Implement**

```go
func Union[T any](a, b *Stream[T]) *Stream[T] {
    return Distinct(FlatMap(Of([][]T{a.Collect(), b.Collect()}), func(s []T) []T { return s }))
}

func UnionE[T any](a, b *Stream[T]) (*Stream[T], error) {
    return Union(a, b), nil
}

func Intersect[T any](a, b *Stream[T]) *Stream[T] {
    bData := b.Collect()
    return Filter(Distinct(a), func(item T) bool {
        return AnyMatch(Of(bData), func(bItem T) bool { return any(item) == any(bItem) })
    })
}

func IntersectE[T any](a, b *Stream[T]) (*Stream[T], error) {
    return Intersect(a, b), nil
}

func Difference[T any](a, b *Stream[T]) *Stream[T] {
    bData := b.Collect()
    return Filter(Distinct(a), func(item T) bool {
        return !AnyMatch(Of(bData), func(bItem T) bool { return any(item) == any(bItem) })
    })
}

func DifferenceE[T any](a, b *Stream[T]) (*Stream[T], error) {
    return Difference(a, b), nil
}
```

**Step 4: Run tests**
Expected: PASS

**Step 5: Commit**

```bash
git add collection/stream.go collection/stream_test.go
git commit -m "feat(collection): add Union, Intersect, Difference set operations"
```

---

## Task 7: IntersectWith Heterogeneous Set Operation

**Files:**
- Modify: `collection/stream.go`
- Modify: `collection/stream_test.go`

**Step 1: Write failing tests**

```go
func TestIntersectWith(t *testing.T) {
    a := Of([]int64{1, 2, 3})
    b := Of([]string{"2", "3", "4"})
    result := IntersectWith(a, b, func(i int64) string { return strconv.FormatInt(i, 10) }).Collect()
    if len(result) != 2 || result[0] != 2 || result[1] != 3 {
        t.Errorf("expected [2 3], got %v", result)
    }
}
```

**Step 2: Run tests**
Expected: FAIL

**Step 3: Implement**

```go
func IntersectWith[T any, U any](a *Stream[T], b *Stream[U], convertFn func(T) U) *Stream[T] {
    bData := b.Collect()
    return Filter(Distinct(a), func(item T) bool {
        converted := convertFn(item)
        return AnyMatch(Of(bData), func(bItem U) bool { return any(converted) == any(bItem) })
    })
}

func IntersectWithE[T any, U any](a *Stream[T], b *Stream[U], convertFn func(T) U) (*Stream[T], error) {
    return IntersectWith(a, b, convertFn), nil
}
```

**Step 4: Run tests**
Expected: PASS

**Step 5: Commit**

```bash
git add collection/stream.go collection/stream_test.go
git commit -m "feat(collection): add IntersectWith heterogeneous set operation"
```

---

## Task 8: GroupBy - Single Key

**Files:**
- Modify: `collection/stream.go`
- Modify: `collection/stream_test.go`

**Step 1: Write failing tests**

```go
type GroupUser struct {
    Dept string
    Name string
}

func TestGroupBy(t *testing.T) {
    users := []GroupUser{{Dept: "A", Name: "a"}, {Dept: "B", Name: "b"}, {Dept: "A", Name: "c"}}
    result := GroupBy(Of(users), func(u GroupUser) string { return u.Dept })
    if len(result) != 2 {
        t.Errorf("expected 2 groups, got %v", len(result))
    }
    if len(result["A"]) != 2 || len(result["B"]) != 1 {
        t.Errorf("unexpected group sizes: %v", result)
    }
}
```

**Step 2: Run tests**
Expected: FAIL

**Step 3: Implement**

```go
func GroupBy[T any, K comparable](s *Stream[T], keyFn func(T) K) map[K][]T {
    result := make(map[K][]T)
    for _, item := range s.Collect() {
        key := keyFn(item)
        result[key] = append(result[key], item)
    }
    return result
}

func GroupByE[T any, K comparable](s *Stream[T], keyFn func(T) K) (map[K][]T, error) {
    return GroupBy(s, keyFn), nil
}
```

**Step 4: Run tests**
Expected: PASS

**Step 5: Commit**

```bash
git add collection/stream.go collection/stream_test.go
git commit -m "feat(collection): add GroupBy single-key grouping"
```

---

## Task 9: GroupByMulti - Composite Key

**Files:**
- Modify: `collection/stream.go`
- Modify: `collection/stream_test.go`

**Step 1: Write failing tests**

```go
func TestGroupByMulti(t *testing.T) {
    users := []GroupUser{{Dept: "A", Name: "a"}, {Dept: "B", Name: "b"}, {Dept: "A", Name: "c"}}
    type key struct{ Dept string }
    result := GroupByMulti(Of(users), func(u GroupUser) key { return key{u.Dept} })
    if len(result) != 2 {
        t.Errorf("expected 2 groups, got %v", len(result))
    }
}
```

**Step 2: Run tests**
Expected: FAIL

**Step 3: Implement**

```go
func GroupByMulti[T any, K any](s *Stream[T], keyFn func(T) K) map[K][]T {
    result := make(map[K][]T)
    for _, item := range s.Collect() {
        key := keyFn(item)
        result[key] = append(result[key], item)
    }
    return result
}

func GroupByMultiE[T any, K any](s *Stream[T], keyFn func(T) K) (map[K][]T, error) {
    return GroupByMulti(s, keyFn), nil
}
```

**Step 4: Run tests**
Expected: PASS

**Step 5: Commit**

```bash
git add collection/stream.go collection/stream_test.go
git commit -m "feat(collection): add GroupByMulti composite-key grouping"
```

---

## Task 10: GroupByWithValue - Value Transformation

**Files:**
- Modify: `collection/stream.go`
- Modify: `collection/stream_test.go`

**Step 1: Write failing tests**

```go
func TestGroupByWithValue(t *testing.T) {
    users := []GroupUser{{Dept: "A", Name: "a"}, {Dept: "B", Name: "b"}, {Dept: "A", Name: "c"}}
    result := GroupByWithValue(Of(users), 
        func(u GroupUser) string { return u.Dept },
        func(u GroupUser) string { return u.Name })
    if len(result) != 2 {
        t.Errorf("expected 2 groups, got %v", len(result))
    }
    if len(result["A"]) != 2 || result["A"][0] != "a" || result["A"][1] != "c" {
        t.Errorf("unexpected group values: %v", result)
    }
}
```

**Step 2: Run tests**
Expected: FAIL

**Step 3: Implement**

```go
func GroupByWithValue[T any, K comparable, V any](s *Stream[T], keyFn func(T) K, valueFn func(T) V) map[K][]V {
    result := make(map[K][]V)
    for _, item := range s.Collect() {
        key := keyFn(item)
        result[key] = append(result[key], valueFn(item))
    }
    return result
}

func GroupByWithValueE[T any, K comparable, V any](s *Stream[T], keyFn func(T) K, valueFn func(T) V) (map[K][]V, error) {
    return GroupByWithValue(s, keyFn, valueFn), nil
}
```

**Step 4: Run tests**
Expected: PASS

**Step 5: Commit**

```bash
git add collection/stream.go collection/stream_test.go
git commit -m "feat(collection): add GroupByWithValue grouping with value transform"
```

---

## Task 11: Final Verification

**Step 1: Run all tests**

Run: `go test ./collection/... -v`
Expected: All pass

**Step 2: Run go vet**

Run: `go vet ./collection/...`
Expected: No issues

**Step 3: Final commit**

```bash
git add -A
git commit -m "feat(collection): complete Stream API extension - conversions, set ops, GroupBy"
```
