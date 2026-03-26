package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestChainMiddleware(t *testing.T) {
	t.Run("chain executes all handlers in order", func(t *testing.T) {
		var executionOrder []int

		handler1 := func(c *gin.Context) {
			executionOrder = append(executionOrder, 1)
			c.Set("key1", "value1")
		}

		handler2 := func(c *gin.Context) {
			executionOrder = append(executionOrder, 2)
			c.Set("key2", "value2")
		}

		handler3 := func(c *gin.Context) {
			executionOrder = append(executionOrder, 3)
			c.Set("key3", "value3")
		}

		chain := ChainMiddleware(handler1, handler2, handler3)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Method: "GET",
			URL:    mustParseURL("/test"),
		}

		chain(c)

		if len(executionOrder) != 3 {
			t.Errorf("expected 3 handlers to execute, got %d", len(executionOrder))
		}

		for i, order := range executionOrder {
			if order != i+1 {
				t.Errorf("expected handler %d at position %d, got %d", i+1, i, order)
			}
		}

		if c.GetString("key1") != "value1" {
			t.Errorf("expected key1=value1, got %s", c.GetString("key1"))
		}
		if c.GetString("key2") != "value2" {
			t.Errorf("expected key2=value2, got %s", c.GetString("key2"))
		}
		if c.GetString("key3") != "value3" {
			t.Errorf("expected key3=value3, got %s", c.GetString("key3"))
		}
	})

	t.Run("chain stops on abort", func(t *testing.T) {
		var executionOrder []int

		handler1 := func(c *gin.Context) {
			executionOrder = append(executionOrder, 1)
		}

		handler2 := func(c *gin.Context) {
			executionOrder = append(executionOrder, 2)
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		handler3 := func(c *gin.Context) {
			executionOrder = append(executionOrder, 3)
		}

		chain := ChainMiddleware(handler1, handler2, handler3)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Method: "GET",
			URL:    mustParseURL("/test"),
		}

		chain(c)

		if len(executionOrder) != 2 {
			t.Errorf("expected 2 handlers to execute (3rd should be skipped), got %d", len(executionOrder))
		}

		if executionOrder[0] != 1 {
			t.Errorf("expected first handler to execute")
		}
		if executionOrder[1] != 2 {
			t.Errorf("expected second handler to execute")
		}

		if c.Writer.Status() != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", c.Writer.Status())
		}
	})

	t.Run("empty chain", func(t *testing.T) {
		chain := ChainMiddleware()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Method: "GET",
			URL:    mustParseURL("/test"),
		}

		// 不应 panic
		chain(c)
	})

	t.Run("single handler", func(t *testing.T) {
		executed := false

		handler := func(c *gin.Context) {
			executed = true
			c.Set("single", "executed")
		}

		chain := ChainMiddleware(handler)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Method: "GET",
			URL:    mustParseURL("/test"),
		}

		chain(c)

		if !executed {
			t.Errorf("expected handler to execute")
		}

		if c.GetString("single") != "executed" {
			t.Errorf("expected single=executed, got %s", c.GetString("single"))
		}
	})

	t.Run("abort in first handler", func(t *testing.T) {
		var executionOrder []int

		handler1 := func(c *gin.Context) {
			executionOrder = append(executionOrder, 1)
			c.AbortWithStatus(http.StatusForbidden)
		}

		handler2 := func(c *gin.Context) {
			executionOrder = append(executionOrder, 2)
		}

		chain := ChainMiddleware(handler1, handler2)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Method: "GET",
			URL:    mustParseURL("/test"),
		}

		chain(c)

		if len(executionOrder) != 1 {
			t.Errorf("expected 1 handler to execute, got %d", len(executionOrder))
		}

		if c.Writer.Status() != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", c.Writer.Status())
		}
	})
}

func TestChainMiddleware_ContextValues(t *testing.T) {
	t.Run("context values are shared between handlers", func(t *testing.T) {
		handler1 := func(c *gin.Context) {
			c.Set("shared", "value_from_handler1")
		}

		handler2 := func(c *gin.Context) {
			val, _ := c.Get("shared")
			if val != "value_from_handler1" {
				t.Errorf("expected shared value from handler1, got %v", val)
			}
			c.Set("shared", "value_from_handler2")
		}

		handler3 := func(c *gin.Context) {
			val, _ := c.Get("shared")
			if val != "value_from_handler2" {
				t.Errorf("expected shared value from handler2, got %v", val)
			}
		}

		chain := ChainMiddleware(handler1, handler2, handler3)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Method: "GET",
			URL:    mustParseURL("/test"),
		}

		chain(c)
	})
}
