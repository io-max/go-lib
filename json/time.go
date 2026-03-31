package json

import (
	"fmt"
	"strings"
	"time"
)

// 支持的时间格式
var timeFormats = []string{
	"2006-01-02 15:04:05",
	"2006-01-02 15:04:05 -0700",
	"2006-01-02 15:04:05Z07:00",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05",
	"2006-01-02",
}

// JsonTime 自定义时间类型，支持多种时间格式解析
type JsonTime struct {
	time.Time
}

// NewJsonTime 创建 JsonTime
func NewJsonTime(t time.Time) JsonTime {
	return JsonTime{Time: t}
}

// NewJsonTimeNow 创建当前时间的 JsonTime
func NewJsonTimeNow() JsonTime {
	return JsonTime{Time: time.Now()}
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (jt *JsonTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")

	// 空值处理
	if s == "" || s == "null" {
		jt.Time = time.Time{}
		return nil
	}

	// 尝试多种格式解析
	for _, format := range timeFormats {
		if t, err := time.ParseInLocation(format, s, time.Local); err == nil {
			jt.Time = t
			return nil
		}
	}

	return fmt.Errorf("invalid time format: %s, supported formats: %v", s, timeFormats)
}

// MarshalJSON 实现 json.Marshaler 接口
func (jt JsonTime) MarshalJSON() ([]byte, error) {
	// 零值处理
	if jt.Time.IsZero() {
		return []byte("null"), nil
	}

	// 统一输出格式为：2006-01-02 15:04:05
	return []byte(fmt.Sprintf("%q", jt.Time.Format("2006-01-02 15:04:05"))), nil
}

// ToTime 转换为 time.Time
func (jt JsonTime) ToTime() time.Time {
	return jt.Time
}

// ToPointer 转换为 *time.Time
func (jt JsonTime) ToPointer() *time.Time {
	if jt.Time.IsZero() {
		return nil
	}
	t := jt.Time
	return &t
}

// IsZero 判断是否为零值
func (jt JsonTime) IsZero() bool {
	return jt.Time.IsZero()
}

// PtrJsonTime 返回 *JsonTime 指针
func PtrJsonTime(jt JsonTime) *JsonTime {
	return &jt
}
