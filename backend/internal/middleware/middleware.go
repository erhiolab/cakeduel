package middleware

import (
	"bufio"
	"cakeduel-backend/internal/logger"
	"cakeduel-backend/internal/utils"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	StatusCode  int
	Size        int
	WroteHeader bool
}

// WriteHeader 捕获状态码
func (rw *responseWriter) WriteHeader(code int) {
	if rw.WroteHeader {
		return
	}
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
	rw.WroteHeader = true
}

// Write 捕获响应体
func (rw *responseWriter) Write(data []byte) (int, error) {
	if !rw.WroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(data)
	rw.Size += n
	return n, err
}

// Flush 支持 http.Flusher
func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack 支持 WebSocket 升级
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hj.Hijack()
}

// RequestID 为每个请求生成请求ID
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.NewString()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(utils.WithRequestID(r.Context(), id)))
	})
}

// Logger 请求日志中间件
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, StatusCode: http.StatusOK}
		next.ServeHTTP(rw, r)
		duration := time.Since(start)
		logger.WithRequestLogCtx(r.Context()).Info("请求",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", rw.StatusCode),
			zap.Int64("duration_ms", duration.Milliseconds()),
			zap.Int("size", rw.Size),
		)
	})
}

// CORS 跨域中间件
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// HealthCheck 健康检查中间件
func HealthCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" || r.URL.Path == "/ping" {
			utils.Success(w, map[string]any{"status": "ok"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
