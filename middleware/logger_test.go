package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

// testMockLogger 用于测试的简单日志记录器
type testMockLogger struct {
	mu    sync.Mutex
	infos []string
}

func (m *testMockLogger) Error(msg string, keys ...interface{}) {
	// 不记录 error
}

func (m *testMockLogger) Info(msg string, keys ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.infos = append(m.infos, msg)
}

func (m *testMockLogger) InfoCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.infos)
}

func (m *testMockLogger) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.infos = nil
}

func TestLoggerMiddleware(t *testing.T) {
	tests := []struct {
		name      string
		config    LoggerMiddlewareConfig
		reqMethod string
		reqPath   string
		wantLog   bool
	}{
		{
			name: "normal request logging",
			config: LoggerMiddlewareConfig{
				SkipPaths: []string{"/health"},
			},
			reqMethod: "GET",
			reqPath:   "/api/users",
			wantLog:   true,
		},
		{
			name: "skip path",
			config: LoggerMiddlewareConfig{
				SkipPaths: []string{"/health"},
			},
			reqMethod: "GET",
			reqPath:   "/health",
			wantLog:   false,
		},
		{
			name: "POST request logging",
			config: LoggerMiddlewareConfig{
				SkipPaths: nil,
			},
			reqMethod: "POST",
			reqPath:   "/api/data",
			wantLog:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &testMockLogger{}

			// 更新配置中的 logger
			tt.config.Logger = logger

			// 创建中间件
			handler := LoggerMiddleware(tt.config)

			// 创建测试上下文
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = &http.Request{
				Method: tt.reqMethod,
				URL:    mustParseURL(tt.reqPath),
			}

			// 设置 request ID
			c.Set(RequestIDKey, "test-trace-id")

			// 执行中间件
			handler(c)

			// 验证日志
			count := logger.InfoCount()
			if tt.wantLog {
				if count == 0 {
					t.Errorf("expected log to be written, but got no logs")
				}
			} else {
				if count > 0 {
					t.Errorf("expected no log for skipped path, but got %d logs", count)
				}
			}
		})
	}
}

func TestLoggerMiddleware_WithNilLogger(t *testing.T) {
	config := LoggerMiddlewareConfig{
		Logger:    nil,
		SkipPaths: []string{},
	}

	handler := LoggerMiddleware(config)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{
		Method: "GET",
		URL:    mustParseURL("/test"),
	}
	c.Set(RequestIDKey, "test-trace")

	// 不应 panic
	handler(c)
}

func TestLoggerMiddleware_LogFields(t *testing.T) {
	logger := &testMockLogger{}

	config := LoggerMiddlewareConfig{
		Logger:    logger,
		SkipPaths: []string{},
	}

	handler := LoggerMiddleware(config)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{
		Method: "POST",
		URL:    mustParseURL("/api/users"),
	}
	c.Set(RequestIDKey, "test-trace-123")

	handler(c)

	count := logger.InfoCount()
	if count != 1 {
		t.Fatalf("expected 1 log entry, got %d", count)
	}
}

func mustParseURL(path string) *url.URL {
	u, err := url.Parse(path)
	if err != nil {
		panic(err)
	}
	return u
}
