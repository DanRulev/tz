package config

import (
	"fmt"
	"time"
	"tz/pkg/valid"

	"github.com/spf13/viper"
)

type DsnConfig struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     string `mapstructure:"port" validate:"required"`
	Username string `mapstructure:"username" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	DBName   string `mapstructure:"db_name" validate:"required"`
	SSLMode  string `mapstructure:"ssl_mode" validate:"required"`
}

type ConnectionConfig struct {
	MaxOpenConns    int           `mapstructure:"max_open_conns" validate:"required"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" validate:"required"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" validate:"required"`
}

type DatabaseConfig struct {
	DsnConfig        DsnConfig        `mapstructure:"dsn"`
	ConnectionConfig ConnectionConfig `mapstructure:"connection"`
}

type LoggerConfig struct {
	Level             string   `mapstructure:"level"`
	Development       bool     `mapstructure:"development"`
	DisableCaller     bool     `mapstructure:"disable_caller"`
	DisableStacktrace bool     `mapstructure:"disable_stacktrace"`
	Encoding          string   `mapstructure:"encoding"`
	OutputPaths       []string `mapstructure:"output_paths"`
	ErrorOutputPaths  []string `mapstructure:"error_output_paths"`
}

type ServerCfg struct {
	Port           string        `mapstructure:"port" validate:"required"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout" validate:"required"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout" validate:"required"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout" validate:"required"`
	MaxHeaderBytes int           `mapstructure:"max_header_bytes" validate:"required"`
}

type Config struct {
	DatabaseConfig DatabaseConfig `mapstructure:"database"`
	LoggerConfig   LoggerConfig   `mapstructure:"logger"`
	Server         ServerCfg      `mapstructure:"server"`
}

func New() (*Config, error) {
	v := viper.New()
	v.AutomaticEnv()

	v.AddConfigPath("./configs")

	name := v.GetString("CONFIG_NAME")
	if name == "" {
		name = "default"
	}

	v.SetConfigName(name)

	bindings := map[string]string{
		"database.dsn.host":     "DB_HOST",
		"database.dsn.port":     "DB_PORT",
		"database.dsn.username": "DB_USERNAME",
		"database.dsn.password": "DB_PASSWORD",
		"database.dsn.db_name":  "DB_NAME",
		"database.dsn.ssl_mode": "DB_SSL",
		"server.port":           "SERVER_PORT",
	}

	for key, env := range bindings {
		if err := v.BindEnv(key, env); err != nil {
			return nil, fmt.Errorf("failed to bind %s to %s: %w", key, env, err)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg := Config{}
	err := v.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := valid.ValidateStruct(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}
