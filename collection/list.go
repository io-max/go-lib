package collection

import (
	"fmt"
	"reflect"
	"strconv"
)

// ToString 将数值类型的元素转换为字符串。
//
//	result := ToString(Of([]int64{1, 2, 3})).Collect()
//	// result: ["1", "2", "3"]
func ToString[T ~int | ~int64 | ~int32 | ~float64 | ~float32](s *Stream[T]) *Stream[string] {
	return Map(s, func(v T) string { return fmt.Sprintf("%v", v) })
}

// ToStringE 返回错误版本（字符串转换不会失败，始终返回 nil error）。
//
//	result, err := ToStringE(Of([]int64{1, 2, 3}))
func ToStringE[T ~int | ~int64 | ~int32 | ~float64 | ~float32](s *Stream[T]) (*Stream[string], error) {
	return ToString(s), nil
}

// StringToInt64 将字符串转换为 int64。转换失败的元素会被设为 0。
//
//	result := StringToInt64(Of([]string{"1", "2", "3"})).Collect()
//	// result: [1, 2, 3]
func StringToInt64(s *Stream[string]) *Stream[int64] {
	return Map(s, func(v string) int64 {
		n, _ := strconv.ParseInt(v, 10, 64)
		return n
	})
}

// StringToInt64E 将字符串转换为 int64，遇到转换错误时返回错误和 nil。
//
//	result, err := StringToInt64E(Of([]string{"1", "a", "3"}))
//	// err: parsing "a": invalid syntax
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

// StringToFloat64 将字符串转换为 float64。转换失败的元素会被设为 0。
//
//	result := StringToFloat64(Of([]string{"1.5", "2.5"})).Collect()
//	// result: [1.5, 2.5]
func StringToFloat64(s *Stream[string]) *Stream[float64] {
	return Map(s, func(v string) float64 {
		f, _ := strconv.ParseFloat(v, 64)
		return f
	})
}

// StringToFloat64E 将字符串转换为 float64，遇到转换错误时返回错误和 nil。
//
//	result, err := StringToFloat64E(Of([]string{"1.5", "a"}))
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

// ToInt 将 int64/int32 转换为 int。
//
//	result := ToInt(Of([]int64{1, 2, 3})).Collect()
//	// result: [1, 2, 3]
func ToInt[T ~int64 | ~int32](s *Stream[T]) *Stream[int] {
	return Map(s, func(v T) int { return int(v) })
}

// ToFloat64 将任意数值类型转换为 float64。
//
//	result := ToFloat64(Of([]int{1, 2, 3})).Collect()
//	// result: [1.0, 2.0, 3.0]
func ToFloat64[T ~int | ~int64 | ~int32 | ~float64 | ~float32](s *Stream[T]) *Stream[float64] {
	return Map(s, func(v T) float64 { return float64(v) })
}

// ToHexString 将整数转换为十六进制字符串。
//
//	result := ToHexString(Of([]int64{255, 16})).Collect()
//	// result: ["ff", "10"]
func ToHexString[T ~int | ~int64 | ~int32](s *Stream[T]) *Stream[string] {
	return Map(s, func(v T) string { return fmt.Sprintf("%x", v) })
}

// ToHexStringE 返回错误版本（十六进制转换不会失败，始终返回 nil error）。
func ToHexStringE[T ~int | ~int64 | ~int32](s *Stream[T]) (*Stream[string], error) {
	return ToHexString(s), nil
}

// ToBytes 将字符串转换为字节切片。
//
//	result := ToBytes(Of([]string{"hello", "world"})).Collect()
//	// result: [[]'h','e','l','l','o'], ['w','o','r','l','d']]
func ToBytes(s *Stream[string]) *Stream[[]byte] {
	return Map(s, func(v string) []byte { return []byte(v) })
}

// ToBytesE 返回错误版本（字节转换不会失败，始终返回 nil error）。
func ToBytesE(s *Stream[string]) (*Stream[[]byte], error) {
	return ToBytes(s), nil
}

// Union 返回两个流的并集，去重。
//
//	result := Union(Of([]int{1, 2, 3}), Of([]int{2, 3, 4})).Collect()
//	// result: [1, 2, 3, 4] (顺序不保证)
func Union[T any](a, b *Stream[T]) *Stream[T] {
	return FlatMap(Of([][]T{a.Collect(), b.Collect()}), func(s []T) []T { return s }).Distinct()
}

// UnionE 返回错误版本（并集操作不会失败，始终返回 nil error）。
func UnionE[T any](a, b *Stream[T]) (*Stream[T], error) {
	return Union(a, b), nil
}

// Intersect 返回两个流的交集，只保留两个流中都存在的元素。
//
//	result := Intersect(Of([]int{1, 2, 3}), Of([]int{2, 3, 4})).Collect()
//	// result: [2, 3]
func Intersect[T any](a, b *Stream[T]) *Stream[T] {
	bData := b.Collect()
	return a.Distinct().Filter(func(item T) bool {
		return Of(bData).AnyMatch(func(bItem T) bool { return reflect.DeepEqual(item, bItem) })
	})
}

// IntersectE 返回错误版本。
func IntersectE[T any](a, b *Stream[T]) (*Stream[T], error) {
	return Intersect(a, b), nil
}

// Difference 返回差集，只保留在 a 中存在但在 b 中不存在的元素。
//
//	result := Difference(Of([]int{1, 2, 3}), Of([]int{2, 3, 4})).Collect()
//	// result: [1]
func Difference[T any](a, b *Stream[T]) *Stream[T] {
	bData := b.Collect()
	return a.Distinct().Filter(func(item T) bool {
		return !Of(bData).AnyMatch(func(bItem T) bool { return reflect.DeepEqual(item, bItem) })
	})
}

// DifferenceE 返回错误版本。
func DifferenceE[T any](a, b *Stream[T]) (*Stream[T], error) {
	return Difference(a, b), nil
}

// IntersectWith 返回异构流的交集，使用 convertFn 将 a 的元素转换为 b 的类型后比较。
//
//	result := IntersectWith(Of([]int64{1, 2, 3}), Of([]string{"2", "3", "4"}),
//		func(i int64) string { return strconv.FormatInt(i, 10) }).Collect()
//	// result: [2, 3]
func IntersectWith[T any, U any](a *Stream[T], b *Stream[U], convertFn func(T) U) *Stream[T] {
	bData := b.Collect()
	return Filter(a.Distinct(), func(item T) bool {
		converted := convertFn(item)
		return Of(bData).AnyMatch(func(bItem U) bool { return reflect.DeepEqual(converted, bItem) })
	})
}

// IntersectWithE 返回错误版本。
func IntersectWithE[T any, U any](a *Stream[T], b *Stream[U], convertFn func(T) U) (*Stream[T], error) {
	return IntersectWith(a, b, convertFn), nil
}

// GroupBy 按 keyFn 提取的键对元素进行分组。
//
//	users := []User{{Dept: "A", Name: "a"}, {Dept: "B", Name: "b"}, {Dept: "A", Name: "c"}}
//	result := GroupBy(Of(users), func(u User) string { return u.Dept })
//	// result: map[string][]User{"A": [{Dept:"A",Name:"a"}, {Dept:"A",Name:"c"}], "B": [{Dept:"B",Name:"b"}]}
func GroupBy[T any, K comparable](s *Stream[T], keyFn func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, item := range s.Collect() {
		key := keyFn(item)
		result[key] = append(result[key], item)
	}
	return result
}

// GroupByE 返回错误版本（分组操作不会失败，始终返回 nil error）。
func GroupByE[T any, K comparable](s *Stream[T], keyFn func(T) K) (map[K][]T, error) {
	return GroupBy(s, keyFn), nil
}

// GroupByMulti 按复合键对元素进行分组。
//
//	type key struct{ Dept string }
//	result := GroupByMulti(Of(users), func(u User) key { return key{u.Dept} })
func GroupByMulti[T any, K comparable](s *Stream[T], keyFn func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, item := range s.Collect() {
		key := keyFn(item)
		result[key] = append(result[key], item)
	}
	return result
}

// GroupByMultiE 返回错误版本。
func GroupByMultiE[T any, K comparable](s *Stream[T], keyFn func(T) K) (map[K][]T, error) {
	return GroupByMulti(s, keyFn), nil
}

// GroupByWithValue 按 keyFn 提取的键对元素进行分组，并对每个元素应用 valueFn 转换为值。
//
//	users := []User{{Dept: "A", Name: "a"}, {Dept: "B", Name: "b"}, {Dept: "A", Name: "c"}}
//	result := GroupByWithValue(Of(users),
//		func(u User) string { return u.Dept },
//		func(u User) string { return u.Name })
//	// result: map[string][]string{"A": ["a", "c"], "B": ["b"]}
func GroupByWithValue[T any, K comparable, V any](s *Stream[T], keyFn func(T) K, valueFn func(T) V) map[K][]V {
	result := make(map[K][]V)
	for _, item := range s.Collect() {
		key := keyFn(item)
		result[key] = append(result[key], valueFn(item))
	}
	return result
}

// GroupByWithValueE 返回错误版本。
func GroupByWithValueE[T any, K comparable, V any](s *Stream[T], keyFn func(T) K, valueFn func(T) V) (map[K][]V, error) {
	return GroupByWithValue(s, keyFn, valueFn), nil
}
