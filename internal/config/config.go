// Package config 提供 Pulse 应用配置管理
package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config 是应用的全局配置
type Config struct {
	// 应用配置
	App AppConfig `mapstructure:"app"`

	// 数据库配置
	Database DatabaseConfig `mapstructure:"database"`

	// 引擎配置
	Engine EngineConfig `mapstructure:"engine"`

	// UI 配置
	UI UIConfig `mapstructure:"ui"`
}

// AppConfig 应用基础配置
type AppConfig struct {
	Name      string `mapstructure:"name"`
	Version   string `mapstructure:"version"`
	DataDir   string `mapstructure:"data_dir"`
	LogLevel  string `mapstructure:"log_level"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	SQLitePath string `mapstructure:"sqlite_path"`
	DuckDBPath  string `mapstructure:"duckdb_path"`
}

// EngineConfig 压测引擎配置
type EngineConfig struct {
	MaxVUs         int  `mapstructure:"max_vus"`
	KeepAlive      bool `mapstructure:"keep_alive"`
	HTTP2          bool `mapstructure:"http2"`
	TimeoutSeconds int  `mapstructure:"timeout_seconds"`
}

// UIConfig UI 配置
type UIConfig struct {
	Theme    string `mapstructure:"theme"`     // dark / light / system
	Language string `mapstructure:"language"`  // zh-CN / en-US
}

// Load 加载配置文件,合并默认值
func Load() (*Config, error) {
	v := viper.New()

	// 默认值
	v.SetDefault("app.name", "Pulse")
	v.SetDefault("app.version", "0.1.0")
	v.SetDefault("app.data_dir", defaultDataDir())
	v.SetDefault("app.log_level", "info")

	v.SetDefault("database.sqlite_path", "${app.data_dir}/pulse.db")
	v.SetDefault("database.duckdb_path", "${app.data_dir}/metrics.duckdb")

	v.SetDefault("engine.max_vus", 20000)
	v.SetDefault("engine.keep_alive", true)
	v.SetDefault("engine.http2", true)
	v.SetDefault("engine.timeout_seconds", 30)

	v.SetDefault("ui.theme", "dark")
	v.SetDefault("ui.language", "zh-CN")

	// 配置文件路径(可选)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("${app.data_dir}")
	v.AddConfigPath(".")

	// 读取配置文件(可选,不存在不报错)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// 环境变量覆盖
	v.SetEnvPrefix("PULSE")
	v.AutomaticEnv()

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// defaultDataDir 返回默认数据目录
func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/pulse"
	}
	return filepath.Join(home, ".pulse", "data")
}