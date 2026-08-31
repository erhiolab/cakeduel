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
	gameHub := hub.NewHub(game.GameConfig{
		RoundsToWin:        gc.RoundsToWin,
		SpecialCardsToAdd:  gc.SpecialCardsToAdd,
		StartingHandLimit:  gc.StartingHandLimit,
		TurnTimeoutSeconds: gc.TurnTimeoutSeconds,
	}, rdb)
	gameHub.SetRoomCodeLen(gc.RoomCodeLength)
	gameHub.SetMatchTimeout(time.Duration(gc.MatchmakingTimeoutMinutes) * time.Minute)
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
