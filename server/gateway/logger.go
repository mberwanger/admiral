package gateway

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.admiral.io/admiral/server/config"
)

func newLogger(cfg *config.Logger) (*zap.Logger, error) {
	return newLoggerWithCore(cfg, nil)
}

func newLoggerWithCore(cfg *config.Logger, zapCore zapcore.Core) (*zap.Logger, error) {
	var c zap.Config
	var opts []zap.Option

	if cfg != nil && cfg.Pretty {
		c = zap.NewDevelopmentConfig()
		opts = append(opts, zap.AddStacktrace(zap.ErrorLevel))
	} else {
		c = zap.NewProductionConfig()
	}

	level := zap.NewAtomicLevel()
	if cfg != nil {
		level.SetLevel(cfg.Level)
	} else {
		level.SetLevel(zapcore.InfoLevel)
	}
	c.Level = level

	logger, err := c.Build(opts...)
	if err != nil {
		return nil, err
	}

	// If zapCore is set, create a new logger, this is currently only used in tests.
	if zapCore != nil {
		logger = zap.New(zapCore, opts...)
	}

	if len(cfg.Namespace) > 0 {
		logger = logger.With(zap.Namespace(cfg.Namespace))
	}

	return logger, nil
}
