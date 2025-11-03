package types

import "time"

// Config 是应用的根配置结构，聚合了所有模块的配置。
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Log      LogConfig      `yaml:"log"`
	System   SystemConfig   `yaml:"system"`
}

// ServerConfig 定义了HTTP服务器的相关配置。
type ServerConfig struct {
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// DatabaseConfig 定义了数据库连接的配置。
type DatabaseConfig struct {
	DSN string `yaml:"dsn"`
}

// RedisConfig 定义了Redis连接的配置。
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// LogConfig 定义了日志记录的相关配置。
type LogConfig struct {
	Level string `yaml:"level"`
	Path  string `yaml:"path"`
}

// SystemConfig 定义了排行榜系统的核心业务配置。
type SystemConfig struct {
	MaxPlayers      int           `yaml:"max_players" env:"MAX_PLAYERS"`
	CacheTTL        time.Duration `yaml:"cache_ttl" env:"CACHE_TTL"`
	CleanupInterval time.Duration `yaml:"cleanup_interval" env:"CLEANUP_INTERVAL"`
	RankUpdateBatch int           `yaml:"rank_update_batch" env:"RANK_UPDATE_BATCH"`
}