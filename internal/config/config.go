package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func Build(file string, envFiles []string, debug bool) *Config {
	tmpLogger := newTmpLogger()

	// Load environment variables from .env files.
	if err := loadEnv(envFiles); err != nil {
		tmpLogger.Fatal("Failed to load environment variables", zap.Error(err))
	}

	// Parse the configuration file.
	cfg, err := parseConfig(file, debug)
	if err != nil {
		tmpLogger.Fatal("Failed to load environment variables", zap.Error(err))
	}

	return cfg
}

func loadEnv(envFiles []string) error {
	// Order is important as godotenv will NOT overwrite existing environment variables.
	for _, filename := range envFiles {
		// Use a temporary logger to parse the environment files
		tmpLogger := newTmpLogger().With(zap.String("file", filename))

		p, err := filepath.Abs(filename)
		if err != nil {
			tmpLogger.Error("failed to get absolute path for .env file", zap.Error(err))
			return err
		}

		// Load the environment variables from the .env file.
		if err := godotenv.Load(p); err != nil {
			// Log a warning if the .env file is not found or cannot be loaded.
			tmpLogger.Warn("Could not load .env file", zap.Error(err))
			continue // Continue loading other files even if one fails.
		}
	}

	return nil
}

func parseConfig(file string, debug bool) (*Config, error) {
	// Use a temporary logger to parse the configuration and output.
	tmpLogger := newTmpLogger().With(zap.String("file", file))

	// Read the configuration file.
	contents, err := os.ReadFile(file)
	if err != nil {
		tmpLogger.Error("failed to read configuration file", zap.Error(err))
		return nil, err
	}

	// Replace environment variables in the configuration content.
	expandedContents := []byte(os.ExpandEnv(string(contents)))

	// Unmarshal the YAML configuration into the config struct.
	cfg := &Config{}
	if err = yaml.Unmarshal(expandedContents, cfg); err != nil {
		tmpLogger.Error("failed to parse configuration file", zap.Error(err))
		return nil, err
	}

	// If debug flag is set, print the configuration and exit.
	if debug {
		b, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			tmpLogger.Fatal("failed to cast configuration file to json", zap.Error(err))
		}
		fmt.Print(string(b))
		os.Exit(0)
	}

	// Set default values in the configuration.
	cfg = setDefaults(cfg)

	// Validate the configuration struct.
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		tmpLogger.Fatal("configuration validation failed", zap.Error(err))
		return nil, err
	}

	return cfg, nil
}

func setDefaults(cfg *Config) *Config {
	if len(cfg.Server.Listener.Address) <= 0 {
		cfg.Server.Listener.Address = "0.0.0.0"
	}
	if cfg.Server.Listener.Port == 0 {
		cfg.Server.Listener.Port = 50051
	}

	if cfg.Server.Logger == nil {
		cfg.Server.Logger = &Logger{Level: zap.ErrorLevel}
	}

	if cfg.Server.Stats == nil {
		cfg.Server.Stats = &Stats{
			FlushInterval: time.Second,
			Prefix:        "admiral",
			ReporterType:  ReporterTypeNull,
		}
	}
	return cfg
}
