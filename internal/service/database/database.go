package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	"github.com/uber-go/tally/v4"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"go.admiral.io/admiral/internal/config"
	"go.admiral.io/admiral/internal/service"
)

const Name = "service.database"

type client struct {
	sqlDB  *sql.DB
	gormDB *gorm.DB
	logger *zap.Logger
	scope  tally.Scope
}

type Client interface {
	DB() *sql.DB
	GormDB() *gorm.DB
}

func (c *client) DB() *sql.DB { return c.sqlDB }

func (c *client) GormDB() *gorm.DB {
	return c.gormDB
}

func New(cfg *config.Config, logger *zap.Logger, scope tally.Scope) (service.Service, error) {
	connection, err := connString(cfg.Services.Database)
	if err != nil {
		return nil, err
	}

	sqlDB, err := sql.Open("postgres", connection)
	if err != nil {
		return nil, err
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &client{sqlDB: sqlDB, gormDB: gormDB, logger: logger, scope: scope}, nil
}

func connString(cfg *config.Database) (string, error) {
	if cfg == nil {
		return "", errors.New("no connection information")
	}

	// Base connection string
	connection := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s",
		cfg.Host, cfg.Port, cfg.DatabaseName, cfg.User, cfg.Password,
	)

	// Handle SSLMode
	switch cfg.SSLMode {
	case config.SSLModeUnspecified, config.SSLModeDisable:
		connection += " sslmode=disable"
	default:
		// Convert to lowercase and replace underscores with hyphens
		mode := strings.ToLower(strings.ReplaceAll(cfg.SSLMode.String(), "_", "-"))
		connection += fmt.Sprintf(" sslmode=%s", mode)
	}

	return connection, nil
}
