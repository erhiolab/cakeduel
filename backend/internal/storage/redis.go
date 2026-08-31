package storage

import (
	"context"
	"cakeduel-backend/internal/config"
	"cakeduel-backend/internal/logger"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// InitRedis 初始化 Redis 连接(可选, 失败时降级为本地缓存)
func InitRedis() *redis.Client {
	cfg := config.Get()
	addr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	rdb := redis.NewClient(&redis.Options{
		Addr:            addr,
		Password:        cfg.Redis.Password,
		DB:              cfg.Redis.DB,
		PoolSize:        cfg.Redis.PoolSize,
		MinIdleConns:    cfg.Redis.MinIdleConnections,
		ConnMaxIdleTime: time.Duration(cfg.Redis.ConnectionMaxIdleTime) * time.Minute,
		DialTimeout:     time.Duration(cfg.Redis.DialTimeout) * time.Second,
		ReadTimeout:     time.Duration(cfg.Redis.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(cfg.Redis.WriteTimeout) * time.Second,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Log.Warn("Redis 连接失败, 降级为本地缓存", zap.Error(err))
		_ = rdb.Close()
		return nil
	}
	logger.Log.Info("Redis 连接成功")
	return rdb
}
