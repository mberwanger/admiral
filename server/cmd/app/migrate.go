package app

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/uber-go/tally/v4"
	"go.uber.org/zap"

	"github.com/mberwanger/admiral/server/config"
	"github.com/mberwanger/admiral/server/service/database"
)

type migrateCmd struct {
	Cmd  *cobra.Command
	opts migrateOpts
}

type migrateOpts struct {
	config string
	force  bool
	down   bool
}

type migrator struct {
	log    *zap.Logger
	config *config.Config
	force  bool
}

//go:embed migrations/*.sql
var fs embed.FS

func newMigrateCmd() *migrateCmd {
	root := &migrateCmd{}

	cmd := &cobra.Command{
		Use:           "migrate",
		Short:         "Database migration tool",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			logger, err := zap.NewDevelopment(zap.AddStacktrace(zap.FatalLevel + 1))
			if err != nil {
				return fmt.Errorf("failed to initialize logger: %w", err)
			}

			cfg := config.Build(configFile, envVarFiles, debug)
			m := &migrator{
				log:    logger,
				config: cfg,
				force:  root.opts.force,
			}

			if root.opts.down {
				return m.Down()
			}
			return m.Up()
		},
	}

	cmd.Flags().BoolVarP(&root.opts.force, "force", "f", false, "Do not ask user for confirmation")
	cmd.Flags().BoolVar(&root.opts.down, "down", false, "Migrates down by one version")

	root.Cmd = cmd
	return root
}

func (m *migrator) setupSqlClient() (*sql.DB, string, error) {
	pgdb, err := database.New(m.config, m.log, tally.NoopScope)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create database client: %w", err)
	}

	dbClient, ok := pgdb.(database.Client)
	if !ok || dbClient.DB() == nil {
		return nil, "", errors.New("no valid database client found")
	}

	hostInfo := fmt.Sprintf("%s@%s:%d",
		m.config.Services.Database.User,
		m.config.Services.Database.Host,
		m.config.Services.Database.Port,
	)
	return dbClient.DB(), hostInfo, nil
}

func (m *migrator) setupSqlMigrator() (*migrate.Migrate, error) {
	sqlDB, _, err := m.setupSqlClient()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	dbDriver, err := postgres.WithInstance(sqlDB, &postgres.Config{
		MigrationsTable: postgres.DefaultMigrationsTable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	srcDriver, err := iofs.New(fs, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to create source driver: %w", err)
	}

	migrator, err := migrate.NewWithInstance("iofs", srcDriver, "postgres", dbDriver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	migrator.Log = &migrateLogger{logger: m.log.Sugar()}
	return migrator, nil
}

func (m *migrator) confirmWithUser(msg string) error {
	_, hostInfo, err := m.setupSqlClient()
	if err != nil {
		return err
	}

	m.log.Info("Using database", zap.String("host", hostInfo))
	if !m.force {
		m.log.Warn(msg)
		fmt.Printf("\n*** Continue with migration? [y/N] ")
		var answer string
		if _, err := fmt.Scanln(&answer); err != nil && !strings.Contains(err.Error(), "unexpected newline") {
			return fmt.Errorf("failed to read user input: %w", err)
		}
		if strings.ToLower(answer) != "y" {
			return errors.New("migration aborted; enter 'y' to continue or use '-f' flag")
		}
		fmt.Println()
	}
	return nil
}

func (m *migrator) Up() error {
	migrator, err := m.setupSqlMigrator()
	if err != nil {
		return err
	}

	msg := "Migration may cause data loss; verify the database details above."
	if err := m.confirmWithUser(msg); err != nil {
		return err
	}

	m.log.Info("Applying up migrations")
	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	m.log.Info("Migrations applied successfully")
	return nil
}

func (m *migrator) Down() error {
	migrator, err := m.setupSqlMigrator()
	if err != nil {
		return err
	}

	version, _, err := migrator.Version()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	msg := fmt.Sprintf(
		"Migrating DOWN by one version (%d -> %d); this may cause data loss.",
		version, version-1,
	)
	if err := m.confirmWithUser(msg); err != nil {
		return err
	}

	m.log.Info("Applying down migration")
	if err := migrator.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply down migration: %w", err)
	}
	m.log.Info("Down migration applied successfully")
	return nil
}

type migrateLogger struct {
	logger *zap.SugaredLogger
}

func (m *migrateLogger) Printf(format string, v ...interface{}) {
	m.logger.Infof(strings.TrimRight(format, "\n"), v...)
}

func (m *migrateLogger) Verbose() bool {
	return true
}
