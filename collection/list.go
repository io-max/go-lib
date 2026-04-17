package collection

import (
	"fmt"
	"reflect"
	"strconv"
)

// ToString 将数值切片转换为字符串切片。
//
//	result := ToString([]int64{1, 2, 3})
//	// result: []string{"1", "2", "3"}
func ToString[T ~int | ~int64 | ~int32 | ~float64 | ~float32](collection []T) []string {
	return Map(Of(collection), func(v T) string { return fmt.Sprintf("%v", v) }).Collect()
}

// StringToInt64 将字符串切片转换为 int64 切片。
//
//	result := StringToInt64([]string{"1", "2", "3"})
//	// result: []int64{1, 2, 3}
func StringToInt64(collection []string) []int64 {
	return Map(Of(collection), func(v string) int64 {
		n, _ := strconv.ParseInt(v, 10, 64)
		return n
	}).Collect()
}

// StringToInt64E 将字符串切片转换为 int64 切片，遇到转换错误时返回错误。
//
//	result, err := StringToInt64E([]string{"1", "a", "3"})
//	// err: parsing "a": invalid syntax
func StringToInt64E(collection []string) ([]int64, error) {
	data := Of(collection).Collect()
	result := make([]int64, len(data))
	for i, v := range data {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		result[i] = n
	}
	return result, nil
}

// StringToFloat64 将字符串切片转换为 float64 切片。
//
//	result := StringToFloat64([]string{"1.5", "2.5"})
//	// result: []float64{1.5, 2.5}
func StringToFloat64(collection []string) []float64 {
	return Map(Of(collection), func(v string) float64 {
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}).Collect()
}

// StringToFloat64E 将字符串切片转换为 float64 切片，遇到转换错误时返回错误。
func StringToFloat64E(collection []string) ([]float64, error) {
	data := Of(collection).Collect()
	result := make([]float64, len(data))
	for i, v := range data {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, err
		}
		result[i] = f
	}
	return result, nil
}

// ToInt 将 int64/int32 切片转换为 int 切片。
//
//	result := ToInt([]int64{1, 2, 3})
//	// result: []int{1, 2, 3}
func ToInt[T ~int64 | ~int32](collection []T) []int {
	return Map(Of(collection), func(v T) int { return int(v) }).Collect()
}

// ToFloat64 将任意数值类型切片转换为 float64 切片。
//
//	result := ToFloat64([]int{1, 2, 3})
//	// result: []float64{1.0, 2.0, 3.0}
func ToFloat64[T ~int | ~int64 | ~int32 | ~float64 | ~float32](collection []T) []float64 {
	return Map(Of(collection), func(v T) float64 { return float64(v) }).Collect()
}

// ToHexString 将整数切片转换为十六进制字符串切片。
//
//	result := ToHexString([]int64{255, 16})
//	// result: []string{"ff", "10"}
func ToHexString[T ~int | ~int64 | ~int32](collection []T) []string {
	return Map(Of(collection), func(v T) string { return fmt.Sprintf("%x", v) }).Collect()
}

// ToBytes 将字符串切片转换为字节切片切片。
//
//	result := ToBytes([]string{"hello", "world"})
//	// result: [][]byte{[]byte("hello"), []byte("world")}
func ToBytes(collection []string) [][]byte {
	return Map(Of(collection), func(v string) []byte { return []byte(v) }).Collect()
}

// Union 返回两个切片的并集，去重。
//
//	result := Union([]int{1, 2, 3}, []int{2, 3, 4})
//	// result: []int{1, 2, 3, 4} (顺序不保证)
func Union[T any](a, b []T) []T {
	return FlatMap(Of([][]T{a, b}), func(s []T) []T { return s }).Distinct().Collect()
}

// Intersect 返回两个切片的交集，只保留两个切片中都存在的元素。
//
//	result := Intersect([]int{1, 2, 3}, []int{2, 3, 4})
//	// result: []int{2, 3}
func Intersect[T any](a, b []T) []T {
	return Of(a).Distinct().Filter(func(item T) bool {
		return Of(b).AnyMatch(func(bItem T) bool { return reflect.DeepEqual(item, bItem) })
	}).Collect()
}

// Difference 返回差集，只保留在 a 中存在但在 b 中不存在的元素。
//
//	result := Difference([]int{1, 2, 3}, []int{2, 3, 4})
//	// result: []int{1}
func Difference[T any](a, b []T) []T {
	return Of(a).Distinct().Filter(func(item T) bool {
		return !Of(b).AnyMatch(func(bItem T) bool { return reflect.DeepEqual(item, bItem) })
	}).Collect()
}

// IntersectWith 返回异构切片的交集，使用 convertFn 将 a 的元素转换为 b 的类型后比较。
//
//	result := IntersectWith([]int64{1, 2, 3}, []string{"2", "3", "4"},
//		func(i int64) string { return strconv.FormatInt(i, 10) })
//	// result: []int64{2, 3}
func IntersectWith[T any, U any](a []T, b []U, convertFn func(T) U) []T {
	return Of(a).Distinct().Filter(func(item T) bool {
		converted := convertFn(item)
		return Of(b).AnyMatch(func(bItem U) bool { return reflect.DeepEqual(converted, bItem) })
	}).Collect()
}

// GroupBy 按 keyFn 提取的键对元素进行分组。
//
//	users := []User{{Dept: "A", Name: "a"}, {Dept: "B", Name: "b"}, {Dept: "A", Name: "c"}}
//	result := GroupBy(users, func(u User) string { return u.Dept })
//	// result: map[string][]User{"A": [{Dept:"A",Name:"a"}, {Dept:"A",Name:"c"}], "B": [{Dept:"B",Name:"b"}]}
func GroupBy[T any, K comparable](collection []T, keyFn func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, item := range Of(collection).Collect() {
		key := keyFn(item)
		result[key] = append(result[key], item)
	}
	return result
}

// GroupByMulti 按复合键对元素进行分组。
//
//	type key struct{ Dept string }
//	result := GroupByMulti(users, func(u User) key { return key{u.Dept} })
func GroupByMulti[T any, K comparable](collection []T, keyFn func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, item := range Of(collection).Collect() {
		key := keyFn(item)
		result[key] = append(result[key], item)
	}
	return result
}

// GroupByWithValue 按 keyFn 提取的键对元素进行分组，并对每个元素应用 valueFn 转换为值。
//
//	users := []User{{Dept: "A", Name: "a"}, {Dept: "B", Name: "b"}, {Dept: "A", Name: "c"}}
//	result := GroupByWithValue(users,
//		func(u User) string { return u.Dept },
//		func(u User) string { return u.Name })
//	// result: map[string][]string{"A": ["a", "c"], "B": ["b"]}
func GroupByWithValue[T any, K comparable, V any](collection []T, keyFn func(T) K, valueFn func(T) V) map[K][]V {
	result := make(map[K][]V)
	for _, item := range Of(collection).Collect() {
		key := keyFn(item)
		result[key] = append(result[key], valueFn(item))
	}
	return result
}
