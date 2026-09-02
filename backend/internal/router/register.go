package router

import (
	"cakeduel-backend/internal/app"
	"cakeduel-backend/internal/controller"
	"net/http"
)

// GameWSHandler WebSocket 处理器
func GameWSHandler(app *app.App) http.HandlerFunc {
	return controller.WebSocketHandler(app)
}

// PingHandler 健康检查
func PingHandler() http.HandlerFunc {
	return controller.Ping()
}
