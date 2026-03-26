package middleware

// JWT 预定义错误码
var (
	JwtErrMissingToken   = NewCode(401001, "Missing token")
	JwtErrInvalidToken   = NewCode(401002, "Invalid token")
	JwtErrTokenBlocked   = NewCode(401003, "Token has been blocked")
	JwtErrUserNotFound   = NewCode(404001, "User not found")
)
