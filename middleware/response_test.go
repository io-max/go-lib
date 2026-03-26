package middleware

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewCode(t *testing.T) {
	err := NewCode(400001, "Invalid parameter")
	assert.Equal(t, 400001, err.Code())
	assert.Equal(t, "Invalid parameter", err.Message())
}

func TestNewCodeDuplicate(t *testing.T) {
	NewCode(500001, "First error")
	assert.Panics(t, func() {
		NewCode(500001, "Duplicate error")
	})
}
