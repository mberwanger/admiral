package gateway

import (
	"github.com/mberwanger/admiral/server/endpoint"
	"github.com/mberwanger/admiral/server/endpoint/application"
	"github.com/mberwanger/admiral/server/endpoint/cluster"
	"github.com/mberwanger/admiral/server/endpoint/healthcheck"
	"github.com/mberwanger/admiral/server/middleware"
	"github.com/mberwanger/admiral/server/middleware/stats"
	"github.com/mberwanger/admiral/server/middleware/validate"
	"github.com/mberwanger/admiral/server/service"
	dbservice "github.com/mberwanger/admiral/server/service/database"
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
