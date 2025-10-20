package db

import (
	"context"
	"fmt"
	"time"
	"tz/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func New(cfg config.DatabaseConfig, log *zap.Logger) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DsnConfig.Host, cfg.DsnConfig.Port, cfg.DsnConfig.Username, cfg.DsnConfig.Password, cfg.DsnConfig.DBName, cfg.DsnConfig.SSLMode)

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(cfg.ConnectionConfig.MaxOpenConns)
	db.SetMaxIdleConns(cfg.ConnectionConfig.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnectionConfig.ConnMaxLifetime)

	return db, nil
}
