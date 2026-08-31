package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Load 加载基础配置
func Load() (*Config, error) {
	configPath := "configs/config.yaml"
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("读取配置文件失败: %v", err)
		return nil, err
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		fmt.Printf("解析配置文件失败: %v", err)
		return nil, err
	}
	// 允许通过环境变量覆盖回合超时(测试/部署用), 避免改动配置文件
	if v := os.Getenv("CAKEDUEL_TURN_TIMEOUT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.Game.TurnTimeoutSeconds = n
		}
	}
	if v := os.Getenv("CAKEDUEL_DISCONNECT_GRACE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.Game.DisconnectGraceSeconds = n
		}
	}
	return &config, nil
}
