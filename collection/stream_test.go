package collection

import "testing"

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
