package json

import (
	"encoding/json"
	"testing"
	"time"
)

func TestJsonTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "datetime with seconds",
			input:    `"2026-10-24 00:00:00"`,
			expected: "2026-10-24 00:00:00",
			wantErr:  false,
		},
		{
			name:     "date only",
			input:    `"2026-10-24"`,
			expected: "2026-10-24 00:00:00",
			wantErr:  false,
		},
		{
			name:     "RFC3339 format",
			input:    `"2026-10-24T00:00:00+08:00"`,
			expected: "2026-10-24 00:00:00",
			wantErr:  false,
		},
		{
			name:     "datetime without T",
			input:    `"2026-10-24T12:30:00"`,
			expected: "2026-10-24 12:30:00",
			wantErr:  false,
		},
		{
			name:     "empty string",
			input:    `""`,
			expected: "",
			wantErr:  false,
		},
		{
			name:    "invalid format",
			input:   `"invalid-date"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jt JsonTime
			err := json.Unmarshal([]byte(tt.input), &jt)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expected != "" {
				// 解析期望时间进行比较（使用本地时区）
				expectedTime, _ := time.ParseInLocation("2006-01-02 15:04:05", tt.expected, time.Local)
				// 比较时忽略时区差异，只比较年月日时分秒
				jtLocal := jt.Time.In(time.Local)
				if jtLocal.Year() != expectedTime.Year() ||
					jtLocal.Month() != expectedTime.Month() ||
					jtLocal.Day() != expectedTime.Day() ||
					jtLocal.Hour() != expectedTime.Hour() ||
					jtLocal.Minute() != expectedTime.Minute() ||
					jtLocal.Second() != expectedTime.Second() {
					t.Errorf("expected %v, got %v", expectedTime, jt.Time)
				}
			}
		})
	}
}

func TestJsonTime_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		time     JsonTime
		expected string
	}{
		{
			name:     "normal time",
			time:     NewJsonTime(time.Date(2026, 10, 24, 12, 30, 0, 0, time.Local)),
			expected: `"2026-10-24 12:30:00"`,
		},
		{
			name:     "zero time",
			time:     JsonTime{},
			expected: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.time)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if string(data) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(data))
			}
		})
	}
}

func TestJsonTime_WithStruct(t *testing.T) {
	type TestStruct struct {
		Name      string   `json:"name"`
		CreatedAt JsonTime `json:"created_at"`
		UpdatedAt JsonTime `json:"updated_at"`
	}

	// 测试反序列化
	jsonStr := `{
		"name": "test",
		"created_at": "2026-10-24 00:00:00",
		"updated_at": "2026-10-25"
	}`

	var ts TestStruct
	err := json.Unmarshal([]byte(jsonStr), &ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", ts.Name)
	}

	// 测试序列化
	data, err := json.Marshal(ts)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	// 验证输出格式
	var ts2 TestStruct
	err = json.Unmarshal(data, &ts2)
	if err != nil {
		t.Fatalf("unexpected unmarshal after marshal: %v", err)
	}
}

func TestJsonTime_ToPointer(t *testing.T) {
	// 非零值
	jt := NewJsonTime(time.Date(2026, 10, 24, 0, 0, 0, 0, time.Local))
	ptr := jt.ToPointer()
	if ptr == nil {
		t.Error("expected non-nil pointer for non-zero time")
	}

	// 零值
	var zero JsonTime
	ptr = zero.ToPointer()
	if ptr != nil {
		t.Error("expected nil pointer for zero time")
	}
}
