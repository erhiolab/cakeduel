package controller

import (
	"cakeduel-backend/internal/app"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocketHandler WebSocket 游戏入口
func WebSocketHandler(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		token := r.URL.Query().Get("token")
		app.Hub.ServeWS(conn, token)
	}
}
