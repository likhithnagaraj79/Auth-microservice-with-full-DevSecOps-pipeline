package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/likhithnagaraj79/auth-service/pkg/config"
	"go.uber.org/zap"
)

func Connect(cfg *config.DatabaseConfig) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Name, cfg.User, cfg.Password, cfg.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	zap.S().Info("connected to PostgreSQL")
	return db, nil
}

func RunMigrations(db *sqlx.DB) error {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
		`CREATE TABLE IF NOT EXISTS users (
			id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			email         VARCHAR(255) UNIQUE NOT NULL,
			username      VARCHAR(50) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL DEFAULT '',
			role          VARCHAR(20) NOT NULL DEFAULT 'user',
			oauth_provider VARCHAR(50) NOT NULL DEFAULT '',
			oauth_id      VARCHAR(255) NOT NULL DEFAULT '',
			is_active     BOOLEAN NOT NULL DEFAULT TRUE,
			created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			last_login_at TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token      VARCHAR(512) UNIQUE NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			revoked    BOOLEAN NOT NULL DEFAULT FALSE
		)`,
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id    UUID REFERENCES users(id) ON DELETE SET NULL,
			action     VARCHAR(100) NOT NULL,
			resource   VARCHAR(255) NOT NULL,
			ip_address VARCHAR(50) NOT NULL DEFAULT '',
			user_agent TEXT NOT NULL DEFAULT '',
			success    BOOLEAN NOT NULL DEFAULT TRUE,
			details    TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	zap.S().Info("database migrations completed")
	return nil
}
