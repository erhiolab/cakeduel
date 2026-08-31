package router

import (
	"cakeduel-backend/internal/app"
	"cakeduel-backend/internal/config"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// New 创建新的路由
func New(app *app.App) *Router {
	return &Router{
		Routes: make(map[string]map[string]http.Handler),
		App:    app,
	}
}

// Register 注册路由
func Register(r *Router, app *app.App) {
	// 游戏 WebSocket
	r.handleFunc(http.MethodGet, "/ws", GameWSHandler(app))

	// 前端静态资源(可选)
	staticPath := config.Get().Gateway.StaticPath
	if staticPath != "" {
		registerStatic(r, staticPath)
	}
}

// registerStatic 注册 SPA 静态资源服务
func registerStatic(r *Router, staticPath string) {
	if _, err := os.Stat(filepath.Join(staticPath, "index.html")); err != nil {
		return
	}
	fileServer := http.FileServer(http.Dir(staticPath))
	r.handle(http.MethodGet, "/static/", http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fileServer.ServeHTTP(w, req)
	})))
	r.handleFunc(http.MethodGet, "/", func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		if path == "/" || strings.HasPrefix(path, "/static/") {
			fileServer.ServeHTTP(w, req)
			return
		}
		// 尝试静态文件, 否则回退到 index.html (SPA)
		full := filepath.Join(staticPath, filepath.Clean("/"+path))
		if fi, err := os.Stat(full); err == nil && !fi.IsDir() {
			fileServer.ServeHTTP(w, req)
			return
		}
		http.ServeFile(w, req, filepath.Join(staticPath, "index.html"))
	})
}
