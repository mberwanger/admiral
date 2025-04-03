package healthcheck

import (
	"context"

	"github.com/uber-go/tally/v4"
	"go.uber.org/zap"

	healthcheckv1 "go.admiral.io/admiral/api/healthcheck/v1"
	"go.admiral.io/admiral/internal/config"
	"go.admiral.io/admiral/internal/endpoint"
)

const Name = "endpoint.healthcheck"

type api struct {
	logger *zap.Logger
	scope  tally.Scope
}

func New(_ *config.Config, logger *zap.Logger, scope tally.Scope) (endpoint.Endpoint, error) {
	return &api{
		logger: logger,
		scope:  scope,
	}, nil
}

func (a *api) Register(r endpoint.Registrar) error {
	healthcheckv1.RegisterHealthcheckAPIServer(r.GRPCServer(), a)
	return r.RegisterJSONGateway(healthcheckv1.RegisterHealthcheckAPIHandler)
}

func (a *api) Healthcheck(context.Context, *healthcheckv1.HealthcheckRequest) (*healthcheckv1.HealthcheckResponse, error) {
	return &healthcheckv1.HealthcheckResponse{}, nil
}
