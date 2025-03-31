package gateway

import (
	"go.admiral.io/admiral/server/endpoint"
	"go.admiral.io/admiral/server/endpoint/application"
	"go.admiral.io/admiral/server/endpoint/cluster"
	"go.admiral.io/admiral/server/endpoint/healthcheck"
	"go.admiral.io/admiral/server/middleware"
	"go.admiral.io/admiral/server/middleware/stats"
	"go.admiral.io/admiral/server/middleware/validate"
	"go.admiral.io/admiral/server/service"
	dbservice "go.admiral.io/admiral/server/service/database"
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
