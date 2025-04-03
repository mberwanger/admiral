package gateway

import (
	"go.admiral.io/admiral/internal/endpoint"
	"go.admiral.io/admiral/internal/endpoint/application"
	"go.admiral.io/admiral/internal/endpoint/cluster"
	"go.admiral.io/admiral/internal/endpoint/healthcheck"
	"go.admiral.io/admiral/internal/middleware"
	"go.admiral.io/admiral/internal/middleware/stats"
	"go.admiral.io/admiral/internal/middleware/validate"
	"go.admiral.io/admiral/internal/service"
	dbservice "go.admiral.io/admiral/internal/service/database"
)

var Services = service.Factory{
	dbservice.Name: dbservice.New,
}

var Middleware = middleware.Factory{
	validate.Name: validate.New,
	stats.Name:    stats.New,
}

var Endpoints = endpoint.Factory{
	application.Name: application.New,
	cluster.Name:     cluster.New,
	healthcheck.Name: healthcheck.New,
}

var CoreComponentFactory = &ComponentFactory{
	Services:   Services,
	Middleware: Middleware,
	Endpoints:  Endpoints,
}
