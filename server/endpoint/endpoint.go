package endpoint

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/uber-go/tally/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/mberwanger/admiral/server/config"
)

type GatewayRegisterAPIHandlerFunc func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error

type Registrar interface {
	GRPCServer() *grpc.Server
	RegisterJSONGateway(GatewayRegisterAPIHandlerFunc) error
}

type Endpoint interface {
	Register(Registrar) error
}

type Factory map[string]func(*config.Config, *zap.Logger, tally.Scope) (Endpoint, error)
