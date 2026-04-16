package collection

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
)

// Stream 提供流式操作支持,支持链式调用和延迟求值。
//
//	nums := []int{1, 2, 3, 4, 5}
//	result := Of(nums).Filter(func(n int) bool { return n%2 == 0 }).Map(func(n int) int { return n * 2 }).Collect()
//	// result: [4, 8]
type Stream[T any] struct {
	executor func() []T
	data     []T
}

// Of 从切片创建 Stream。nil 输入会被转换为空切片。
//
//	s := Of([]int{1, 2, 3})
//	s := Of([]int(nil)) // 安全的空流
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

// Collect 收集流中的所有元素到切片并返回。这是一个终端操作。
//
//	s := Of([]int{1, 2, 3})
//	arr := s.Collect() // []int{1, 2, 3}
func (s *Stream[T]) Collect() []T {
	if s.executor != nil {
		s.data = s.executor()
		s.executor = nil
	}
	result := make([]T, len(s.data))
	copy(result, s.data)
	return result
}

// Filter 过滤流中的元素,只保留满足 predicate 的元素。支持链式调用。
//
//	nums := []int{1, 2, 3, 4, 5}
//	evens := Of(nums).Filter(func(n int) bool { return n%2 == 0 }).Collect()
//	// evens: [2, 4]
//
//	// 独立函数形式
//	evens := Filter(Of(nums), func(n int) bool { return n%2 == 0 }).Collect()
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

// Filter 独立函数形式,支持函数组合风格。
//
//	evens := Filter(Of([]int{1, 2, 3, 4}), func(n int) bool { return n%2 == 0 }).Collect()
//	// evens: [2, 4]
func Filter[T any](s *Stream[T], pred func(T) bool) *Stream[T] {
	return s.Filter(pred)
}

// Map 对流中的每个元素执行转换操作,返回包含转换后元素的新流。
//
//	nums := []int{1, 2, 3}
//	doubled := Map(Of(nums), func(n int) int { return n * 2 }).Collect()
//	// doubled: [2, 4, 6]
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

// FlatMap 对流中的每个元素执行转换,将返回的切片展开合并到结果流中。
// 常用于扁平化嵌套结构。
//
//	nested := [][]int{{1, 2}, {3, 4}, {5}}
//	flat := FlatMap(Of(nested), func(s []int) []int { return s }).Collect()
//	// flat: [1, 2, 3, 4, 5]
//
//	// 将字符串列表按字符拆分
//	words := []string{"hello", "world"}
//	chars := FlatMap(Of(words), func(s string) []rune { return []rune(s) }).Collect()
//	// chars: ['h','e','l','l','o','w','o','r','l','d']
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

// Distinct 去重,移除流中重复的元素。使用 reflect.DeepEqual 进行相等性判断。
//
//	nums := []int{1, 2, 2, 3, 3, 3}
//	unique := Of(nums).Distinct().Collect()
//	// unique: [1, 2, 3]
//
//	// 自定义去重规则
//	users := []User{{ID: 1, Name: "a"}, {ID: 1, Name: "b"}, {ID: 2, Name: "c"}}
//	uniqueUsers := Distinct(Of(users), func(a, b User) bool { return a.ID == b.ID }).Collect()
//	// uniqueUsers: [{ID: 1, Name: "a"}, {ID: 2, Name: "c"}]
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

// Distinct 使用自定义相等函数进行去重。
//
//	users := []User{{ID: 1, Name: "a"}, {ID: 1, Name: "b"}, {ID: 2, Name: "c"}}
//	unique := Distinct(Of(users), func(a, b User) bool { return a.ID == b.ID }).Collect()
//	// unique: [{ID: 1, Name: "a"}, {ID: 2, Name: "c"}]
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

// Peek 对流中的每个元素执行副作用操作,同时保持元素不变通过流。
// 常用于调试,观察流经管道的元素。
//
//	var nums []int
//	Of([]int{1, 2, 3}).Peek(func(n int) { nums = append(nums, n) }).Collect()
//	// nums: [1, 2, 3]
//
//	// 调试用,观察过滤前的数据
//	Of([]int{1, 2, 3, 4}).Peek(func(n int) { fmt.Printf("before filter: %d\n", n) }).
//		Filter(func(n int) bool { return n%2 == 0 }).
//		Peek(func(n int) { fmt.Printf("after filter: %d\n", n) }).
//		Collect()
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

// Count 返回流中元素的数量。这是一个终端操作。
//
//	count := Of([]int{1, 2, 3, 4, 5}).Count()
//	// count: 5
func (s *Stream[T]) Count() int {
	return len(s.Collect())
}

// Reduce 对流中的所有元素进行聚合,使用 initial 作为初始值。
// accumulator 函数接收当前累计值和下一个元素,返回新的累计值。
//
//	sum := Of([]int{1, 2, 3, 4}).Reduce(0, func(acc, n int) int { return acc + n })
//	// sum: 10
//
//	// 计算最大值
//	max := Of([]int{3, 1, 4, 1, 5, 9, 2, 6}).Reduce(math.MinInt, func(acc, n int) int {
//	    if n > acc { return n }
//	    return acc
//	})
//	// max: 9
//
//	// 字符串连接
//	joined := Of([]string{"hello", " ", "world"}).Reduce("", func(acc, s string) string { return acc + s })
//	// joined: "hello world"
func (s *Stream[T]) Reduce(initial T, fn func(T, T) T) T {
	result := initial
	for _, item := range s.Collect() {
		result = fn(result, item)
	}
	return result
}

// FindFirst 返回流中的第一个元素。如果流为空,返回 (零值, false)。
//
//	first, ok := Of([]int{1, 2, 3}).FindFirst()
//	// first: 1, ok: true
//
//	first, ok := Of([]int{}).FindFirst()
//	// first: 0, ok: false
//
//	// 配合 Filter 使用
//	first, ok := Of([]int{1, 2, 3}).Filter(func(n int) bool { return n > 1 }).FindFirst()
//	// first: 2, ok: true
func (s *Stream[T]) FindFirst() (T, bool) {
	data := s.Collect()
	if len(data) == 0 {
		return *new(T), false
	}
	return data[0], true
}

// AnyMatch 判断是否存在任意元素满足 predicate。存在时短路返回 true。
//
//	hasEven := Of([]int{1, 2, 3}).AnyMatch(func(n int) bool { return n%2 == 0 })
//	// hasEven: true
//
//	hasNegative := Of([]int{1, 2, 3}).AnyMatch(func(n int) bool { return n < 0 })
//	// hasNegative: false
func (s *Stream[T]) AnyMatch(pred func(T) bool) bool {
	for _, item := range s.Collect() {
		if pred(item) {
			return true
		}
	}
	return false
}

// AllMatch 判断是否所有元素都满足 predicate。存在不满足元素时短路返回 false。
// 空流返回 true( vacuous truth )。
//
//	allPositive := Of([]int{1, 2, 3}).AllMatch(func(n int) bool { return n > 0 })
//	// allPositive: true
//
//	allEven := Of([]int{1, 2, 3}).AllMatch(func(n int) bool { return n%2 == 0 })
//	// allEven: false
//
//	empty := Of([]int{}).AllMatch(func(n int) bool { return n > 0 })
//	// empty: true
func (s *Stream[T]) AllMatch(pred func(T) bool) bool {
	for _, item := range s.Collect() {
		if !pred(item) {
			return false
		}
	}
	return true
}

// NoneMatch 判断是否所有元素都不满足 predicate。与 AllMatch 逻辑相反。
// 存在满足元素时返回 false。空流返回 true。
//
//	noneNegative := Of([]int{1, 2, 3}).NoneMatch(func(n int) bool { return n < 0 })
//	// noneNegative: true
//
//	noneZero := Of([]int{1, 0, 2}).NoneMatch(func(n int) bool { return n == 0 })
//	// noneZero: false
func (s *Stream[T]) NoneMatch(pred func(T) bool) bool {
	return !s.AnyMatch(pred)
}

// Limit 截取流的前 n 个元素,超过的部分被丢弃。
// n <= 0 时返回空流。
//
//	limited := Of([]int{1, 2, 3, 4, 5}).Limit(3).Collect()
//	// limited: [1, 2, 3]
//
//	zero := Of([]int{1, 2, 3}).Limit(0).Collect()
//	// zero: []
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

// Skip 跳过流的前 n 个元素,返回剩余元素组成的流。
// n <= 0 时返回原流的副本。
//
//	skipped := Of([]int{1, 2, 3, 4, 5}).Skip(2).Collect()
//	// skipped: [3, 4, 5]
//
//	skipMore := Of([]int{1, 2, 3}).Skip(10).Collect()
//	// skipMore: []
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

// Sorted 对流中的元素进行排序。排序是即时求值的,因为需要完整数据。
//
//	sorted := Of([]int{3, 1, 2}).Sorted(func(a, b int) bool { return a < b }).Collect()
//	// sorted: [1, 2, 3]
//
//	// 降序排列
//	desc := Of([]int{3, 1, 2}).Sorted(func(a, b int) bool { return a > b }).Collect()
//	// desc: [3, 2, 1]
//
//	// 按字段排序
//	users := []User{{Age: 30}, {Age: 20}, {Age: 40}}
//	sortedUsers := Of(users).Sorted(func(a, b User) bool { return a.Age < b.Age }).Collect()
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

// ForEach 对流中的每个元素执行 fn。通常用于产生副作用。
// 这是一个终端操作。
//
//	var sum int
//	Of([]int{1, 2, 3}).ForEach(func(n int) { sum += n })
//	// sum: 6
//
//	// 打印所有元素
//	Of([]string{"a", "b", "c"}).ForEach(func(s string) { fmt.Println(s) })
func (s *Stream[T]) ForEach(fn func(T)) {
	for _, item := range s.Collect() {
		fn(item)
	}
}

// ToMap 将流转换为 map。使用 keyFn 提取每个元素的 key,value 为元素本身。
// 如果 keyFn 产生重复 key,后面的值会覆盖前面的。
//
//	users := []User{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}}
//	m := ToMap(Of(users), func(u User) int64 { return u.ID })
//	// m: map[int64]User{1: {ID:1, Name:"a"}, 2: {ID:2, Name:"b"}}
func ToMap[T any, K comparable](s *Stream[T], keyFn func(T) K) map[K]T {
	result := make(map[K]T)
	for _, item := range s.Collect() {
		result[keyFn(item)] = item
	}
	return result
}

// ToMapWithValue 将流转换为 map。使用 keyFn 提取 key,valueFn 提取 value。
// 如果 keyFn 产生重复 key,后面的值会覆盖前面的。
//
//	users := []User{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}}
//	m := ToMapWithValue(Of(users),
//		func(u User) int64 { return u.ID },
//		func(u User) string { return u.Name })
//	// m: map[int64]string{1: "a", 2: "b"}
func ToMapWithValue[T any, K comparable, V any](s *Stream[T], keyFn func(T) K, valueFn func(T) V) map[K]V {
	result := make(map[K]V)
	for _, item := range s.Collect() {
		result[keyFn(item)] = valueFn(item)
	}
	return result
}

// ToSet 将流转换为 set(使用 map[T]struct{} 实现)。
//
//	nums := []int{1, 2, 2, 3, 3, 3}
//	set := ToSet(Of(nums))
//	// set: map[int]struct{}{1: {}, 2: {}, 3: {}}
//	_, has2 := set[2] // true
func ToSet[T comparable](s *Stream[T]) map[T]struct{} {
	result := make(map[T]struct{})
	for _, item := range s.Collect() {
		result[item] = struct{}{}
	}
	return result
}

func ToString[T ~int | ~int64 | ~int32 | ~float64 | ~float32](s *Stream[T]) *Stream[string] {
	return Map(s, func(v T) string { return fmt.Sprintf("%v", v) })
}

func ToStringE[T ~int | ~int64 | ~int32 | ~float64 | ~float32](s *Stream[T]) (*Stream[string], error) {
	return ToString(s), nil
}

func StringToInt64(s *Stream[string]) *Stream[int64] {
	return Map(s, func(v string) int64 {
		n, _ := strconv.ParseInt(v, 10, 64)
		return n
	})
}

func StringToInt64E(s *Stream[string]) (*Stream[int64], error) {
	data := s.Collect()
	result := make([]int64, len(data))
	for i, v := range data {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		result[i] = n
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
	data := s.Collect()
	result := make([]float64, len(data))
	for i, v := range data {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, err
		}
		result[i] = f
	}
	return Of(result), nil
}

func ToInt[T ~int64 | ~int32](s *Stream[T]) *Stream[int] {
	return Map(s, func(v T) int { return int(v) })
}

func ToFloat64[T ~int | ~int64 | ~int32 | ~float64 | ~float32](s *Stream[T]) *Stream[float64] {
	return Map(s, func(v T) float64 { return float64(v) })
}
