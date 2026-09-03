package controller

import (
	"cakeduel-backend/internal/app"
	"cakeduel-backend/internal/utils"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// ReplayShareHandler 保存分享回放, 返回访问 ID 与链接(Redis 存 1 天)
func ReplayShareHandler(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(io.LimitReader(r.Body, 4*1024*1024+1024))
		if err != nil {
			utils.BadRequest(w, "读取回放数据失败")
			return
		}
		// 简单结构校验: 必须是可解析对象且含 frames
		var probe map[string]any
		if err := json.Unmarshal(raw, &probe); err != nil {
			utils.BadRequest(w, "回放数据格式错误")
			return
		}
		if _, ok := probe["frames"]; !ok {
			utils.BadRequest(w, "回放数据缺少帧")
			return
		}
		id, err := app.Hub.ShareReplay(raw)
		if err != nil {
			utils.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.Success(w, map[string]any{
			"id":  id,
			"url": "/replay/" + id,
		})
	}
}

// ReplayShareGetHandler 按 ID 读取分享回放(公开)
func ReplayShareGetHandler(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/replay/")
		id = strings.Trim(id, "/")
		if id == "" {
			utils.NotFound(w, "缺少回放 ID")
			return
		}
		raw, ok := app.Hub.FetchSharedReplay(id)
		if !ok {
			utils.Error(w, http.StatusNotFound, "回放不存在或已过期")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(raw))
	}
}
