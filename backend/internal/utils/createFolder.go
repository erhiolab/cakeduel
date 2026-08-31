package utils

import (
	"cakeduel-backend/internal/config"
	"os"
)

// CreateFolder 创建运行所需目录
func CreateFolder() {
	cfg := config.Get()
	for _, path := range []string{cfg.Gateway.DataPath, cfg.Gateway.TempPath, "logs"} {
		if path == "" {
			continue
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			_ = os.MkdirAll(path, 0o755)
		}
	}
}
