# Collection 工具类设计

## 概述

设计一个流式操作的集合工具类，参考 Java Stream API，提供 `Map`、`Filter`、`Chunk` 等链式调用能力。

## 核心类型

```go
package collection

type Stream[T any] struct {
    data []T
}

func Of[T any](collection []T) *Stream[T]
```

## API 设计

| 方法 | 功能 |
|------|------|
| `Of[T]([]T)` | 从切片创建 Stream |
| `Map[U any](fn func(T) U)` | 类型转换，返回新的 Stream[U] |
| `Filter(fn func(T) bool)` | 过滤元素 |
| `Chunk(size int)` | 按固定大小拆分成多个 Stream |
| `Collect()` | 收集结果到切片 |
| `ToMap[K comparable](fn func(T) K)` | 收集到 Map |
| `First()` (T, bool) | 获取第一个元素 |
| `Find(fn func(T) bool)` (T, bool) | 查找满足条件的第一个元素 |
| `Count()` int | 元素数量 |
| `IsEmpty()` bool | 是否为空 |

## 使用示例

```go
// 提取 ID 并过滤
ids := collection.Of(users).
    Map(func(u User) int64 { return u.ID }).
    Filter(func(id int64) bool { return id > 0 }).
    Collect()

// 拆分大列表为小块
chunks := collection.Of(items).Chunk(100)
for _, chunk := range chunks {
    // 处理每块
}
```

## 文件结构

```
collection/
    stream.go      # Stream[T] 核心实现
    stream_test.go # 单元测试
```

## 后续扩展

- `Reduce` 聚合操作
- `FlatMap` 扁平化映射
- `GroupBy` 分组
