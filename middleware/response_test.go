package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewCode(t *testing.T) {
	err := NewCode(800001, "Test error")
	assert.Equal(t, 800001, err.Code())
	assert.Equal(t, "Test error", err.Message())
}

func TestNewCodeDuplicate(t *testing.T) {
	NewCode(999001, "First error")
	assert.Panics(t, func() {
		NewCode(999001, "Duplicate error")
	})
}

func TestRespondSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set(RequestIDKey, "test-trace-id-123")

	RespondSuccess(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":200`)
	assert.Contains(t, w.Body.String(), `"message":"success"`)
	assert.Contains(t, w.Body.String(), `"trace_id":"test-trace-id-123"`)
}

func TestRespondSuccessWithData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set(RequestIDKey, "test-trace-id-456")

	data := map[string]interface{}{
		"user_id":   123,
		"username":  "testuser",
		"is_active": true,
	}
	RespondSuccessWithData(c, data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":200`)
	assert.Contains(t, w.Body.String(), `"message":"success"`)
	assert.Contains(t, w.Body.String(), `"user_id":123`)
	assert.Contains(t, w.Body.String(), `"username":"testuser"`)
}

func TestRespondSuccessWithMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set(RequestIDKey, "test-trace-id-789")

	data := map[string]interface{}{"status": "ok"}
	RespondSuccessWithMessage(c, data, "Operation completed")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":200`)
	assert.Contains(t, w.Body.String(), `"message":"Operation completed"`)
	assert.Contains(t, w.Body.String(), `"status":"ok"`)
}

func TestRespondErrorWithCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set(RequestIDKey, "test-trace-id-err-001")

	errCode := NewCode(400801, "Test Bad Request Error")
	RespondErrorWithCode(c, errCode)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"code":400801`)
	assert.Contains(t, w.Body.String(), `"message":"Test Bad Request Error"`)
	assert.Contains(t, w.Body.String(), `"trace_id":"test-trace-id-err-001"`)
}

func TestRespondErrorWithMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set(RequestIDKey, "test-trace-id-err-002")

	errCode := NewCode(400802, "Original Error")
	customMessage := "Custom error message"
	RespondErrorWithMessage(c, errCode, customMessage)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"code":400802`)
	assert.Contains(t, w.Body.String(), `"message":"Custom error message"`)
}

func TestRespondErrorWithFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set(RequestIDKey, "test-trace-id-err-003")

	errCode := NewCode(400803, "Template Error")
	RespondErrorWithFormat(c, errCode, "User %s not found, ID: %d", "john", 42)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"code":400803`)
	assert.Contains(t, w.Body.String(), `"message":"User john not found, ID: 42"`)
}

func TestRespondErrorWithHTTPStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set(RequestIDKey, "test-trace-id-err-004")

	errCode := NewCode(400804, "Resource Not Found")
	RespondErrorWithHTTPStatus(c, http.StatusNotFound, errCode)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), `"code":400804`)
	assert.Contains(t, w.Body.String(), `"message":"Resource Not Found"`)
}

func TestRespondErrorWithDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set(RequestIDKey, "test-trace-id-err-005")

	originalErr := errors.New("database connection failed")
	errCode := NewCode(500801, "Internal Server Error")
	RespondErrorWithDetails(c, originalErr, errCode)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"code":500801`)
	assert.Contains(t, w.Body.String(), `"message":"Internal Server Error"`)
	assert.Contains(t, w.Body.String(), `"error":"database connection failed"`)
}

func TestRespondErrorWithHTTPStatusAndMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set(RequestIDKey, "test-trace-id-err-006")

	RespondErrorWithHTTPStatusAndMessage(c, http.StatusServiceUnavailable, "Service Temporarily Unavailable")

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), `"code":503000`)
	assert.Contains(t, w.Body.String(), `"message":"Service Temporarily Unavailable"`)
}

func TestHTTPStatusFromCode(t *testing.T) {
	tests := []struct {
		code     int
		expected int
	}{
		{400001, http.StatusBadRequest},
		{400999, http.StatusBadRequest},
		{401000, http.StatusUnauthorized},
		{402999, http.StatusUnauthorized},
		{403000, http.StatusForbidden},
		{403999, http.StatusForbidden},
		{404000, http.StatusNotFound},
		{428999, http.StatusNotFound},
		{429000, http.StatusTooManyRequests},
		{499999, http.StatusTooManyRequests},
		{500000, http.StatusInternalServerError},
		{500001, http.StatusInternalServerError},
		{999999, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("code_%d", tt.code), func(t *testing.T) {
			result := httpStatusFromCode(tt.code)
			assert.Equal(t, tt.expected, result)
		})
	}
}
