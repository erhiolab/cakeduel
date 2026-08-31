package utils

import (
	"context"
	"math/rand"
)

// RequestIDKey 请求ID在context中的键
type requestIDKey struct{}

// RequestIDKey 请求ID键
var RequestIDKey = requestIDKey{}

// WithRequestID 将请求ID写入context
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, RequestIDKey, id)
}

// RandomString 生成随机字符串
func RandomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
