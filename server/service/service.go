package service

import (
	"github.com/mberwanger/admiral/server/config"
	"github.com/uber-go/tally/v4"
	"go.uber.org/zap"
)

type Service interface{}

type Factory map[string]func(*config.Config, *zap.Logger, tally.Scope) (Service, error)

var Registry = map[string]Service{}
