package service

import (
	"github.com/uber-go/tally/v4"
	"go.uber.org/zap"
	
	"go.admiral.io/admiral/internal/config"
)

type Service interface{}

type Factory map[string]func(*config.Config, *zap.Logger, tally.Scope) (Service, error)

var Registry = map[string]Service{}
