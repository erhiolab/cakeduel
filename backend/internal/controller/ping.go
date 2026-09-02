package controller

import (
	"cakeduel-backend/internal/utils"
	"cakeduel-backend/internal/version"
	"net/http"
	"time"
)

// Ping 服务健康检查
func Ping() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		utils.Success(w, map[string]any{
			"status":      "ok",
			"service":     "cakeduel",
			"version":     version.Version,
			"serverTime":  time.Now().UnixMilli(),
		})
	}
}
