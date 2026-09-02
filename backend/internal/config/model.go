package config

// gatewayConfig 网关配置
type gatewayConfig struct {
	Port             int    `yaml:"port"`
	DataPath         string `yaml:"data-path"`
	TempPath         string `yaml:"temp-path"`
	StaticPath       string `yaml:"static-path"`
	LocalCacheExpire int    `yaml:"local-cache-expire"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Output         string `yaml:"output"`
	Level          string `yaml:"level"`
	LogPath        string `yaml:"log-path"`
	RequestLogPath string `yaml:"request-log-path"`
	MaxSize        int    `yaml:"max-size"`
	MaxBackups     int    `yaml:"max-backups"`
	MaxAge         int    `yaml:"max-age"`
	Compress       bool   `yaml:"compress"`
}

// redisConfig Redis配置
type redisConfig struct {
	Host                  string `yaml:"host"`
	Port                  int    `yaml:"port"`
	Password              string `yaml:"password"`
	DB                    int    `yaml:"db"`
	PoolSize              int    `yaml:"pool-size"`
	ProjectPrefix         string `yaml:"project-prefix"`
	MinIdleConnections    int    `yaml:"min-idle-connections"`
	ConnectionMaxIdleTime int    `yaml:"connection-max-idle-time"`
	DialTimeout           int    `yaml:"dial-timeout"`
	ReadTimeout           int    `yaml:"read-timeout"`
	WriteTimeout          int    `yaml:"write-timeout"`
}

// gameConfig 游戏配置
type gameConfig struct {
	RoundsToWin               int `yaml:"rounds-to-win"`
	SpecialCardsToAdd         int `yaml:"special-cards-to-add"`
	StartingHandLimit         int `yaml:"starting-hand-limit"`
	RoomCodeLength            int    `yaml:"room-code-length"`
	MatchmakingTimeoutSeconds int    `yaml:"matchmaking-timeout-seconds"`
	MatchmakingStorage        string `yaml:"matchmaking-storage"`
	TurnTimeoutSeconds        int    `yaml:"turn-timeout-seconds"`
	DisconnectGraceSeconds    int    `yaml:"disconnect-grace-seconds"`
}

// Config 配置
type Config struct {
	Gateway gatewayConfig `yaml:"gateway"`
	Logger  LoggerConfig  `yaml:"logger"`
	Redis   redisConfig   `yaml:"redis"`
	Game    gameConfig    `yaml:"game"`
}
