package controller

import (
	"cakeduel-backend/internal/app"
	"cakeduel-backend/internal/utils"
	"net/http"
)

// RoomsHandler 公开房间列表(主界面一键观战)
func RoomsHandler(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		utils.Success(w, map[string]any{
			"rooms": app.Hub.PublicRooms(),
		})
	}
}
