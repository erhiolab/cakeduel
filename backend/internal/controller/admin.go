package controller

import (
	"cakeduel-backend/internal/app"
	"cakeduel-backend/internal/utils"
	"encoding/json"
	"net/http"
	"strings"
)

// AdminPageHandler 管理员页面入口: 每次访问生成一个一分钟有效的密码
func AdminPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(adminPageHTML))
	}
}

// AdminChallengeHandler 生成管理员访问密码
func AdminChallengeHandler(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if err := app.Hub.AdminCreateChallenge(); err != nil {
			utils.Error(w, http.StatusInternalServerError, "密码生成失败")
			return
		}
		utils.Success(w, map[string]any{
			"created":   true,
			"expiresIn": 60,
		})
	}
}

// AdminVerifyHandler 校验密码并发放管理员令牌
func AdminVerifyHandler(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			utils.BadRequest(w, "请求格式错误")
			return
		}
		token, err := app.Hub.AdminVerify(body.Password)
		if err != nil {
			utils.Error(w, http.StatusUnauthorized, err.Error())
			return
		}
		utils.Success(w, map[string]any{
			"token": token,
		})
	}
}

// AdminOverviewHandler 在线用户与房间概况
func AdminOverviewHandler(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizeAdmin(app, w, r) {
			return
		}
		utils.Success(w, map[string]any{
			"online":    app.Hub.AdminClients(),
			"roomCount": len(app.Hub.AdminRooms()),
		})
	}
}

// AdminRoomsHandler 房间详情(双方手牌/聊天/观战)
func AdminRoomsHandler(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizeAdmin(app, w, r) {
			return
		}
		rooms := app.Hub.AdminRooms()
		utils.Success(w, map[string]any{"rooms": rooms})
	}
}

// AdminDismissHandler 强制解散房间
func AdminDismissHandler(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizeAdmin(app, w, r) {
			return
		}
		var body struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Code) == "" {
			utils.BadRequest(w, "缺少房间码")
			return
		}
		if err := app.Hub.AdminDismissRoom(strings.ToUpper(strings.TrimSpace(body.Code))); err != nil {
			utils.Error(w, http.StatusNotFound, err.Error())
			return
		}
		utils.Success(w, map[string]any{"dismissed": true})
	}
}

// AdminSettingsHandler 读取后台设置(创建房间开关)
func AdminSettingsHandler(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizeAdmin(app, w, r) {
			return
		}
		utils.Success(w, map[string]any{
			"creationEnabled": app.Hub.CreationEnabled(),
		})
	}
}

// AdminSettingsUpdateHandler 更新后台设置(关闭/开启创建房间)
func AdminSettingsUpdateHandler(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizeAdmin(app, w, r) {
			return
		}
		var body struct {
			CreationEnabled bool `json:"creationEnabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			utils.BadRequest(w, "请求格式错误")
			return
		}
		app.Hub.SetCreationEnabled(body.CreationEnabled)
		utils.Success(w, map[string]any{
			"creationEnabled": app.Hub.CreationEnabled(),
		})
	}
}

// authorizeAdmin 校验 Bearer 管理员令牌
func authorizeAdmin(app *app.App, w http.ResponseWriter, r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	token := strings.TrimPrefix(auth, "Bearer ")
	if !app.Hub.AdminAuthorize(token) {
		utils.Error(w, http.StatusUnauthorized, "未授权或登录已过期")
		return false
	}
	return true
}
