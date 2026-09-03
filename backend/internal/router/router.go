package router

import (
	"cakeduel-backend/internal/app"
	"cakeduel-backend/internal/config"
	"cakeduel-backend/internal/controller"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// New 创建新的路由
func New(app *app.App) *Router {
	return &Router{
		Routes:   make(map[string]map[string]http.Handler),
		Prefixes: make(map[string]map[string]http.Handler),
		App:      app,
	}
}

// Register 注册路由
func Register(r *Router, app *app.App) {
	// 健康检查
	r.handleFunc(http.MethodGet, "/ping", PingHandler())

	// 公开房间列表(主界面观战入口)
	r.handleFunc(http.MethodGet, "/api/rooms", controller.RoomsHandler(app))

	// 分享回放: POST /api/replay/share, GET /api/replay/{id}
	r.handleFunc(http.MethodPost, "/api/replay/share", controller.ReplayShareHandler(app))
	r.handlePrefixFunc(http.MethodGet, "/api/replay/", controller.ReplayShareGetHandler(app))

	// 游戏 WebSocket
	r.handleFunc(http.MethodGet, "/ws", GameWSHandler(app))

	// 管理员页面与接口
	r.handleFunc(http.MethodGet, "/admin", controller.AdminPageHandler())
	r.handleFunc(http.MethodGet, "/api/admin/challenge", controller.AdminChallengeHandler(app))
	r.handleFunc(http.MethodPost, "/api/admin/verify", controller.AdminVerifyHandler(app))
	r.handleFunc(http.MethodGet, "/api/admin/overview", controller.AdminOverviewHandler(app))
	r.handleFunc(http.MethodGet, "/api/admin/rooms", controller.AdminRoomsHandler(app))
	r.handleFunc(http.MethodPost, "/api/admin/rooms/dismiss", controller.AdminDismissHandler(app))
	r.handleFunc(http.MethodGet, "/api/admin/settings", controller.AdminSettingsHandler(app))
	r.handleFunc(http.MethodPost, "/api/admin/settings", controller.AdminSettingsUpdateHandler(app))

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
