package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// TestJwtClaims 测试用 Claims
type TestJwtClaims struct {
	jwt.RegisteredClaims
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}

// TestJwtValidator 测试用验证器
type TestJwtValidator struct {
	validateFunc func(c *gin.Context, claims jwt.Claims) bool
	getUserFunc  func(c *gin.Context, claims jwt.Claims) bool
}

func (v *TestJwtValidator) ValidateToken(c *gin.Context, claims jwt.Claims) bool {
	if v.validateFunc != nil {
		return v.validateFunc(c, claims)
	}
	return true
}

func (v *TestJwtValidator) GetUser(c *gin.Context, claims jwt.Claims) bool {
	if v.getUserFunc != nil {
		return v.getUserFunc(c, claims)
	}
	return true
}

func createTestToken(secret []byte, claims *TestJwtClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func setupTestRouter(path string, handler gin.HandlerFunc, testHandler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(handler)
	if testHandler != nil {
		router.GET(path, testHandler)
	} else {
		router.GET(path, func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}
	return router
}

func TestJwtAuth_MissingToken(t *testing.T) {
	cfg := JwtConfig{
		Secret: []byte("test-secret"),
	}

	handler := JwtAuth(cfg, func() jwt.Claims { return &TestJwtClaims{} })
	router := setupTestRouter("/test1", handler, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Missing token")
}

func TestJwtAuth_InvalidToken(t *testing.T) {
	cfg := JwtConfig{
		Secret: []byte("test-secret"),
	}

	handler := JwtAuth(cfg, func() jwt.Claims { return &TestJwtClaims{} })
	router := setupTestRouter("/test2", handler, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test2", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid token")
}

func TestJwtAuth_ValidToken(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Now()

	claims := &TestJwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:   123,
		Username: "testuser",
	}

	tokenStr, err := createTestToken(secret, claims)
	assert.NoError(t, err)

	cfg := JwtConfig{
		Secret: secret,
	}

	handler := JwtAuth(cfg, func() jwt.Claims { return &TestJwtClaims{} })
	router := setupTestRouter("/test3", handler, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test3", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestJwtAuth_CustomTokenHeader(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Now()

	claims := &TestJwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:   123,
		Username: "testuser",
	}

	tokenStr, err := createTestToken(secret, claims)
	assert.NoError(t, err)

	cfg := JwtConfig{
		Secret:      secret,
		TokenHeader: "X-Custom-Token",
		TokenPrefix: "!", // 禁用前缀
	}

	handler := JwtAuth(cfg, func() jwt.Claims { return &TestJwtClaims{} })
	router := setupTestRouter("/test4", handler, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test4", nil)
	req.Header.Set("X-Custom-Token", tokenStr)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJwtAuth_Validator(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Now()

	claims := &TestJwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:   123,
		Username: "testuser",
	}

	tokenStr, err := createTestToken(secret, claims)
	assert.NoError(t, err)

	// Test with custom validator that passes
	validator := &TestJwtValidator{
		validateFunc: func(c *gin.Context, claims jwt.Claims) bool {
			return true
		},
		getUserFunc: func(c *gin.Context, claims jwt.Claims) bool {
			c.Set("user_id", claims.(*TestJwtClaims).UserID)
			return true
		},
	}

	cfg := JwtConfig{
		Secret:    secret,
		Validator: validator,
	}

	handler := JwtAuth(cfg, func() jwt.Claims { return &TestJwtClaims{} })
	router := setupTestRouter("/test5", handler, func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test5", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "123")
}

func TestJwtAuth_Validator_BlockedToken(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Now()

	claims := &TestJwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:   123,
		Username: "testuser",
	}

	tokenStr, err := createTestToken(secret, claims)
	assert.NoError(t, err)

	// Test with validator that blocks the token
	validator := &TestJwtValidator{
		validateFunc: func(c *gin.Context, claims jwt.Claims) bool {
			return false // Block this token
		},
	}

	cfg := JwtConfig{
		Secret:    secret,
		Validator: validator,
	}

	handler := JwtAuth(cfg, func() jwt.Claims { return &TestJwtClaims{} })
	router := setupTestRouter("/test6", handler, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test6", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Token has been blocked")
}

func TestJwtAuth_Validator_UserNotFound(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Now()

	claims := &TestJwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:   123,
		Username: "testuser",
	}

	tokenStr, err := createTestToken(secret, claims)
	assert.NoError(t, err)

	// Test with validator that cannot find user
	validator := &TestJwtValidator{
		validateFunc: func(c *gin.Context, claims jwt.Claims) bool {
			return true
		},
		getUserFunc: func(c *gin.Context, claims jwt.Claims) bool {
			return false // User not found
		},
	}

	cfg := JwtConfig{
		Secret:    secret,
		Validator: validator,
	}

	handler := JwtAuth(cfg, func() jwt.Claims { return &TestJwtClaims{} })
	router := setupTestRouter("/test7", handler, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test7", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "User not found")
}

func TestGetJwtClaims(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Now()

	claims := &TestJwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   "123",
		},
		UserID:   123,
		Username: "testuser",
	}

	tokenStr, err := createTestToken(secret, claims)
	assert.NoError(t, err)

	cfg := JwtConfig{
		Secret: secret,
	}

	handler := JwtAuth(cfg, func() jwt.Claims { return &TestJwtClaims{} })
	router := setupTestRouter("/test8", handler, func(c *gin.Context) {
		retrievedClaims := GetJwtClaims[*TestJwtClaims](c)
		c.JSON(http.StatusOK, gin.H{
			"user_id":  retrievedClaims.UserID,
			"username": retrievedClaims.Username,
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test8", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "123")
	assert.Contains(t, w.Body.String(), "testuser")
}

func TestJwtAuth_ExpiredToken(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Now()

	// Create expired token
	claims := &TestJwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now.Add(-time.Hour * 2)),
		},
		UserID:   123,
		Username: "testuser",
	}

	tokenStr, err := createTestToken(secret, claims)
	assert.NoError(t, err)

	cfg := JwtConfig{
		Secret: secret,
	}

	handler := JwtAuth(cfg, func() jwt.Claims { return &TestJwtClaims{} })
	router := setupTestRouter("/test9", handler, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test9", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid token")
}

func TestJwtAuth_WrongSecret(t *testing.T) {
	secret := []byte("test-secret")
	wrongSecret := []byte("wrong-secret")
	now := time.Now()

	claims := &TestJwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:   123,
		Username: "testuser",
	}

	// Create token with one secret
	tokenStr, err := createTestToken(secret, claims)
	assert.NoError(t, err)

	// Try to validate with different secret
	cfg := JwtConfig{
		Secret: wrongSecret,
	}

	handler := JwtAuth(cfg, func() jwt.Claims { return &TestJwtClaims{} })
	router := setupTestRouter("/test10", handler, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test10", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid token")
}

func TestJwtAuth_NoPrefix(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Now()

	claims := &TestJwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:   123,
		Username: "testuser",
	}

	tokenStr, err := createTestToken(secret, claims)
	assert.NoError(t, err)

	cfg := JwtConfig{
		Secret:      secret,
		TokenPrefix: "!", // 禁用前缀
	}

	handler := JwtAuth(cfg, func() jwt.Claims { return &TestJwtClaims{} })
	router := setupTestRouter("/test11", handler, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test11", nil)
	req.Header.Set("Authorization", tokenStr) // No Bearer prefix
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJwtAuth_CustomSigningMethod(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Now()

	claims := &TestJwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:   123,
		Username: "testuser",
	}

	// Create token with HS512
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString(secret)
	assert.NoError(t, err)

	cfg := JwtConfig{
		Secret:        secret,
		SigningMethod: jwt.SigningMethodHS512,
	}

	handler := JwtAuth(cfg, func() jwt.Claims { return &TestJwtClaims{} })
	router := setupTestRouter("/test12", handler, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test12", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetJwtClaims_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test13", func(c *gin.Context) {
		claims := GetJwtClaims[*TestJwtClaims](c)
		if claims == nil {
			c.JSON(http.StatusOK, gin.H{"claims": nil})
		} else {
			c.JSON(http.StatusOK, gin.H{"claims": claims})
		}
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test13", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestValidatorWithError 测试验证器返回 error 的情况
func TestValidatorWithError(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Now()

	claims := &TestJwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:   123,
		Username: "testuser",
	}

	tokenStr, err := createTestToken(secret, claims)
	assert.NoError(t, err)

	// Validator that returns an error
	validator := &TestJwtValidator{
		validateFunc: func(c *gin.Context, claims jwt.Claims) bool {
			return true
		},
		getUserFunc: func(c *gin.Context, claims jwt.Claims) bool {
			// Simulate error when getting user
			return false
		},
	}

	cfg := JwtConfig{
		Secret:    secret,
		Validator: validator,
	}

	handler := JwtAuth(cfg, func() jwt.Claims { return &TestJwtClaims{} })
	router := setupTestRouter("/test14", handler, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test14", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "User not found")
}

// TestJwtAuth_NilClaimsFunc 测试 nil claimsFunc 的情况
func TestJwtAuth_NilClaimsFunc(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Now()

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(now),
		Subject:   "test-subject",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(secret)
	assert.NoError(t, err)

	cfg := JwtConfig{
		Secret: secret,
	}

	// nil claimsFunc should use default RegisteredClaims
	handler := JwtAuth(cfg, nil)
	router := setupTestRouter("/test15", handler, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test15", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
