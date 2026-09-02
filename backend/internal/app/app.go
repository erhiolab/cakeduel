package app

import (
	"cakeduel-backend/internal/config"
	"cakeduel-backend/internal/game"
	"cakeduel-backend/internal/hub"
	"cakeduel-backend/internal/logger"
	"cakeduel-backend/internal/storage"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// App 应用实例
type App struct {
	Redis   *redis.Client
	Hub     *hub.Hub
	Enabled bool
}

// New 创建新的应用实例
func New() *App {
	// 初始化 Redis (可选)
	rdb := storage.InitRedis()

	// 初始化游戏房间中心
	gc := config.Get().Game
	// 注意: InitRedis 失败时返回 nil 指针, 赋给接口会变成"非 nil 的空接口",
	// 这里显式转成真正的 nil 接口, 避免 hub 中 h.rdb == nil 判断失效
	var cmdable redis.Cmdable
	if rdb != nil {
		cmdable = rdb
	}
	gameHub := hub.NewHub(game.GameConfig{
		RoundsToWin:        gc.RoundsToWin,
		SpecialCardsToAdd:  gc.SpecialCardsToAdd,
		StartingHandLimit:  gc.StartingHandLimit,
		TurnTimeoutSeconds: gc.TurnTimeoutSeconds,
	}, cmdable)
	gameHub.SetRoomCodeLen(gc.RoomCodeLength)
	gameHub.SetMatchTimeout(time.Duration(gc.MatchmakingTimeoutSeconds) * time.Second)
	gameHub.SetDisconnectGrace(time.Duration(gc.DisconnectGraceSeconds) * time.Second)

	return &App{
		Redis: rdb,
		Hub:   gameHub,
	}
}

// Close 关闭应用实例的所有资源
func (app *App) Close() {
	app.Hub.Close()
	logger.Log.Info("关闭游戏房间中心")

	if app.Redis != nil {
		logger.Log.Info("关闭 Redis")
		if err := app.Redis.Close(); err != nil {
			logger.Log.Warn("关闭 Redis 连接失败", zap.Error(err))
		}
	}
}
