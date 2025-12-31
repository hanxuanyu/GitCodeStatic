package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Web       WebConfig       `yaml:"web"`
	Workspace WorkspaceConfig `yaml:"workspace"`
	Storage   StorageConfig   `yaml:"storage"`
	Worker    WorkerConfig    `yaml:"worker"`
	Cache     CacheConfig     `yaml:"cache"`
	Security  SecurityConfig  `yaml:"security"`
	Git       GitConfig       `yaml:"git"`
	Log       LogConfig       `yaml:"log"`
	Metrics   MetricsConfig   `yaml:"metrics"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// WebConfig 前端配置
type WebConfig struct {
	Dir     string `yaml:"dir"`
	Enabled bool   `yaml:"enabled"`
}

// WorkspaceConfig 工作空间配置
type WorkspaceConfig struct {
	BaseDir  string `yaml:"base_dir"`
	CacheDir string `yaml:"cache_dir"`
	StatsDir string `yaml:"stats_dir"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type     string         `yaml:"type"` // sqlite/postgres
	SQLite   SQLiteConfig   `yaml:"sqlite"`
	Postgres PostgresConfig `yaml:"postgres"`
}

// SQLiteConfig SQLite配置
type SQLiteConfig struct {
	Path string `yaml:"path"`
}

// PostgresConfig PostgreSQL配置
type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"sslmode"`
}

// WorkerConfig Worker配置
type WorkerConfig struct {
	CloneWorkers   int `yaml:"clone_workers"`
	PullWorkers    int `yaml:"pull_workers"`
	StatsWorkers   int `yaml:"stats_workers"`
	GeneralWorkers int `yaml:"general_workers"`
	QueueBuffer    int `yaml:"queue_buffer"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	MaxTotalSize    int64 `yaml:"max_total_size"`
	MaxSingleResult int64 `yaml:"max_single_result"`
	RetentionDays   int   `yaml:"retention_days"`
	CleanupInterval int   `yaml:"cleanup_interval"` // seconds
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	EncryptionKey string `yaml:"encryption_key"`
}

// GitConfig Git配置
type GitConfig struct {
	CommandPath     string `yaml:"command_path"`
	FallbackToGoGit bool   `yaml:"fallback_to_gogit"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `yaml:"level"`  // debug/info/warn/error
	Format string `yaml:"format"` // json/text
	Output string `yaml:"output"` // stdout/file path
}

// MetricsConfig 指标配置
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 从环境变量覆盖敏感配置
	if key := os.Getenv("ENCRYPTION_KEY"); key != "" {
		cfg.Security.EncryptionKey = key
	}

	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		cfg.Storage.SQLite.Path = dbPath
	}

	// 设置默认值
	setDefaults(&cfg)

	return &cfg, nil
}

// setDefaults 设置默认值
func setDefaults(cfg *Config) {
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 30 * time.Second
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 30 * time.Second
	}

	if cfg.Workspace.BaseDir == "" {
		cfg.Workspace.BaseDir = "./workspace"
	}
	if cfg.Workspace.CacheDir == "" {
		cfg.Workspace.CacheDir = "./workspace/cache"
	}
	if cfg.Workspace.StatsDir == "" {
		cfg.Workspace.StatsDir = "./workspace/stats"
	}

	if cfg.Storage.Type == "" {
		cfg.Storage.Type = "sqlite"
	}
	if cfg.Storage.SQLite.Path == "" {
		cfg.Storage.SQLite.Path = "./workspace/data.db"
	}

	if cfg.Worker.CloneWorkers == 0 {
		cfg.Worker.CloneWorkers = 2
	}
	if cfg.Worker.PullWorkers == 0 {
		cfg.Worker.PullWorkers = 2
	}
	if cfg.Worker.StatsWorkers == 0 {
		cfg.Worker.StatsWorkers = 2
	}
	if cfg.Worker.GeneralWorkers == 0 {
		cfg.Worker.GeneralWorkers = 4
	}
	if cfg.Worker.QueueBuffer == 0 {
		cfg.Worker.QueueBuffer = 100
	}

	if cfg.Cache.MaxTotalSize == 0 {
		cfg.Cache.MaxTotalSize = 10 * 1024 * 1024 * 1024 // 10GB
	}
	if cfg.Cache.MaxSingleResult == 0 {
		cfg.Cache.MaxSingleResult = 100 * 1024 * 1024 // 100MB
	}
	if cfg.Cache.RetentionDays == 0 {
		cfg.Cache.RetentionDays = 30
	}
	if cfg.Cache.CleanupInterval == 0 {
		cfg.Cache.CleanupInterval = 3600 // 1 hour
	}

	if cfg.Git.FallbackToGoGit {
		// Default: allow fallback
	}

	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.Format == "" {
		cfg.Log.Format = "json"
	}
	if cfg.Log.Output == "" {
		cfg.Log.Output = "stdout"
	}

	if cfg.Metrics.Path == "" {
		cfg.Metrics.Path = "/metrics"
	}
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	cfg := &Config{}
	setDefaults(cfg)
	return cfg
}
