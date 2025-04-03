package config

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBuild(t *testing.T) {
	configContent := `
server:
  listener:
    address: "127.0.0.1"
    port: ${LISTENER_PORT}
`
	envContent := "LISTENER_PORT=8888"

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	envFile := filepath.Join(tmpDir, ".env")

	err := os.WriteFile(configFile, []byte(configContent), 0600)
	assert.NoError(t, err)
	err = os.WriteFile(envFile, []byte(envContent), 0600)
	assert.NoError(t, err)

	t.Run("ValidConfigAndEnv", func(t *testing.T) {
		cfg := Build(configFile, []string{envFile}, false)
		assert.NotNil(t, cfg)
		assert.Equal(t, "127.0.0.1", cfg.Server.Listener.Address)
		assert.Equal(t, "8888", os.Getenv("LISTENER_PORT"))
		assert.Equal(t, 8888, cfg.Server.Listener.Port)
	})
}

func TestLoadEnv(t *testing.T) {
	tmpDir := t.TempDir()
	envFile1 := filepath.Join(tmpDir, ".env1")
	envFile2 := filepath.Join(tmpDir, ".env2")

	err := os.WriteFile(envFile1, []byte("KEY1=value1"), 0600)
	assert.NoError(t, err)
	err = os.WriteFile(envFile2, []byte("KEY2=value2"), 0600)
	assert.NoError(t, err)

	t.Run("LoadMultipleEnvFiles", func(t *testing.T) {
		err := loadEnv([]string{envFile1, envFile2})
		assert.NoError(t, err)
		assert.Equal(t, "value1", os.Getenv("KEY1"))
		assert.Equal(t, "value2", os.Getenv("KEY2"))
	})

	t.Run("NonExistentEnvFile", func(t *testing.T) {
		err := loadEnv([]string{envFile1, "nonexistent.env"})
		assert.NoError(t, err)
		assert.Equal(t, "value1", os.Getenv("KEY1"))
	})
}

func TestParseConfig(t *testing.T) {
	// Setup: Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	// Test case 1: Valid config file
	t.Run("ValidConfig", func(t *testing.T) {
		content := `
server:
  listener:
    address: "0.0.0.0"
    port: 50051
`
		err := os.WriteFile(configFile, []byte(content), 0600)
		assert.NoError(t, err)

		cfg, err := parseConfig(configFile, false)
		assert.NoError(t, err)
		assert.Equal(t, "0.0.0.0", cfg.Server.Listener.Address)
		assert.Equal(t, 50051, cfg.Server.Listener.Port)
	})

	// Test case 2: Invalid YAML
	t.Run("InvalidYAML", func(t *testing.T) {
		content := `
server:
  listener:
    address: "0.0.0.0"
    port: invalid # Invalid type
`
		err := os.WriteFile(configFile, []byte(content), 0600)
		assert.NoError(t, err)

		_, err = parseConfig(configFile, false)
		assert.Error(t, err)
	})
}

func TestSetDefaults(t *testing.T) {
	// Test case 1: Empty config, apply defaults
	t.Run("ApplyDefaults", func(t *testing.T) {
		cfg := &Config{}
		cfg = setDefaults(cfg)
		assert.Equal(t, "0.0.0.0", cfg.Server.Listener.Address)
		assert.Equal(t, 50051, cfg.Server.Listener.Port)
		assert.NotNil(t, cfg.Server.Logger)
		assert.Equal(t, zap.ErrorLevel, cfg.Server.Logger.Level)
		assert.NotNil(t, cfg.Server.Stats)
		assert.Equal(t, time.Second, cfg.Server.Stats.FlushInterval)
		assert.Equal(t, "admiral", cfg.Server.Stats.Prefix)
		assert.Equal(t, ReporterTypeNull, cfg.Server.Stats.ReporterType)
	})

	// Test case 2: Partial config, preserve values and apply defaults
	t.Run("PartialConfig", func(t *testing.T) {
		cfg := &Config{
			Server: Server{
				Listener: Listener{
					Address: "127.0.0.1",
				},
			},
		}
		cfg = setDefaults(cfg)
		assert.Equal(t, "127.0.0.1", cfg.Server.Listener.Address)
		assert.Equal(t, 50051, cfg.Server.Listener.Port)
		assert.NotNil(t, cfg.Server.Logger)
		assert.NotNil(t, cfg.Server.Stats)
	})
}
