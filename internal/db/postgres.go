package db

import (
	"context"
	"database/sql"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"time"

	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
)

type Config struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type AppConfig struct {
	DB     Config       `yaml:"db"`
	Server ServerConfig `yaml:"server"`
}

func LoadConfigFromYAML(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var appConfig AppConfig
	if err := yaml.Unmarshal(data, &appConfig); err != nil {
		return Config{}, err
	}

	return appConfig.DB, nil
}

// InitConnection initializes and verifies a secure DB connection
func InitConnection(cfg Config, logger *zap.Logger) *sql.DB {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Fatal("Failed to open DB", zap.Error(err))
	}

	// Connection pool settings (fine-tune per use case)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logger.Fatal("DB ping failed", zap.Error(err))
	}

	logger.Info("Successfully connected to database")
	createTables(db, logger)
	return db
}

func createTables(db *sql.DB, logger *zap.Logger) {
	schema := `
	CREATE TABLE IF NOT EXISTS books (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		author TEXT NOT NULL,
		isbn TEXT NOT NULL
	);`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, schema); err != nil {
		logger.Fatal("Failed to create books table", zap.Error(err))
	}
}
