package collection

import "testing"

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
