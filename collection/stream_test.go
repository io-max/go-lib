package collection

import (
	"fmt"
	"testing"
)

type User struct {
	ID   int64
	Name string
}

func TestOf(t *testing.T) {
	// nil 输入
	s1 := Of([]int(nil))
	if len(s1.Collect()) != 0 {
		t.Errorf("expected empty slice, got %v", s1.Collect())
	}

	// 正常输入
	s2 := Of([]int{1, 2, 3})
	if len(s2.Collect()) != 3 {
		t.Errorf("expected 3 elements, got %v", s2.Collect())
	}

	// 验证实际内容
	result := s2.Collect()
	for i, v := range []int{1, 2, 3} {
		if result[i] != v {
			t.Errorf("expected result[%d]=%d, got %d", i, v, result[i])
		}
	}
}

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

func TestMap(t *testing.T) {
	users := []User{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}}
	ids := Map(Of(users), func(u User) int64 { return u.ID }).Collect()

	if len(ids) != 2 || ids[0] != 1 || ids[1] != 2 {
		t.Errorf("expected [1 2], got %v", ids)
	}
}

func TestFlatMap(t *testing.T) {
	nested := [][]int{{1, 2}, {3, 4}, {5}}
	flat := FlatMap(Of(nested), func(s []int) []int { return s }).Collect()
	if len(flat) != 5 || flat[0] != 1 || flat[4] != 5 {
		t.Errorf("expected [1 2 3 4 5], got %v", flat)
	}
}

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

func TestPeek(t *testing.T) {
	var peeked []int
	nums := []int{1, 2, 3}
	Of(nums).Peek(func(n int) { peeked = append(peeked, n) }).Collect()
	if len(peeked) != 3 || peeked[0] != 1 || peeked[2] != 3 {
		t.Errorf("expected [1 2 3], got %v", peeked)
	}
}

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

func TestForEach(t *testing.T) {
	var sum int
	Of([]int{1, 2, 3}).ForEach(func(n int) { sum += n })
	if sum != 6 {
		t.Errorf("expected 6, got %d", sum)
	}
}

func TestToMap(t *testing.T) {
	users := []User{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}}
	m := ToMap(Of(users), func(u User) int64 { return u.ID })
	if len(m) != 2 || m[1].Name != "a" {
		t.Errorf("expected map with 2 entries, got %v", m)
	}
}

func TestToMapWithValue(t *testing.T) {
	users := []User{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}}
	m := ToMapWithValue(Of(users), func(u User) int64 { return u.ID }, func(u User) string { return u.Name })
	if len(m) != 2 || m[1] != "a" || m[2] != "b" {
		t.Errorf("expected map with 2 entries, got %v", m)
	}
}

func TestToSet(t *testing.T) {
	nums := []int{1, 2, 2, 3, 3, 3}
	s := ToSet(Of(nums))
	if len(s) != 3 {
		t.Errorf("expected set with 3 entries, got %v", s)
	}
	if _, ok := s[1]; !ok {
		t.Errorf("expected 1 in set")
	}
}

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

func TestStringToFloat64(t *testing.T) {
	result := StringToFloat64(Of([]string{"1.5", "2.5"})).Collect()
	if len(result) != 2 || result[0] != 1.5 || result[1] != 2.5 {
		t.Errorf("expected [1.5 2.5], got %v", result)
	}
}

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

func TestToHexString(t *testing.T) {
	result := ToHexString(Of([]int64{255, 16})).Collect()
	if len(result) != 2 || result[0] != "ff" || result[1] != "10" {
		t.Errorf("expected [ff 10], got %v", result)
	}
}

func TestToBytes(t *testing.T) {
	result := ToBytes(Of([]string{"hello", "world"})).Collect()
	if len(result) != 2 || string(result[0]) != "hello" {
		t.Errorf("expected [hello world], got %v", result)
	}
}

func TestConvert(t *testing.T) {
	result := Convert(Of([]int{1, 2, 3}), func(i int) string {
		return fmt.Sprintf("num_%d", i)
	}).Collect()
	expected := []string{"num_1", "num_2", "num_3"}
	if len(result) != 3 || result[0] != "num_1" {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

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
