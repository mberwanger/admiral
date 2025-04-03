package cluster

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uber-go/tally/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	clusterv1 "go.admiral.io/admiral/api/cluster/v1"
	"go.admiral.io/admiral/internal/config"
	"go.admiral.io/admiral/internal/endpoint"
	"go.admiral.io/admiral/internal/model"
	"go.admiral.io/admiral/internal/service"
	"go.admiral.io/admiral/internal/service/database"
)

const Name = "endpoint.cluster"

type api struct {
	logger *zap.Logger
	scope  tally.Scope
	sqlDb  *sql.DB
	gormDB *gorm.DB
	//queryBuilder *querybuilder.QueryBuilder
}

func New(_ *config.Config, log *zap.Logger, scope tally.Scope) (endpoint.Endpoint, error) {
	db, ok := service.Registry["service.database"]
	if !ok {
		return nil, errors.New("could not find db service")
	}

	dbClient, ok := db.(database.Client)
	if !ok {
		return nil, errors.New("service was not the correct type")
	}

	api := &api{
		logger: log,
		scope:  scope,
		sqlDb:  dbClient.DB(),
		gormDB: dbClient.GormDB(),
		//queryBuilder: querybuilder.New([]string{"name", "password_login_enabled", "oauth2_login_enabled", "saml2_login_enabled"}),
	}
	return api, nil
}

func (a *api) Register(r endpoint.Registrar) error {
	clusterv1.RegisterClusterAPIServer(r.GRPCServer(), a)
	return r.RegisterJSONGateway(clusterv1.RegisterClusterAPIHandler)
}

func (a *api) CreateCluster(ctx context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.CreateClusterResponse, error) {
	cluster := model.Cluster{
		Name: req.GetName(),
	}

	result := a.gormDB.WithContext(ctx).Create(&cluster)
	if result.Error != nil {
		return nil, result.Error
	}

	return &clusterv1.CreateClusterResponse{Cluster: model.ConvertClusterToProto(&cluster)}, nil
}

func (a *api) ListClusters(ctx context.Context, req *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (a *api) GetCluster(ctx context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
	var cluster model.Cluster

	result := a.gormDB.WithContext(ctx).First(&cluster, "id = ?", req.GetId())
	if e := result.Error; e != nil {
		if errors.Is(e, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "cluster not found")
		} else {
			return nil, status.Error(codes.Internal, result.Error.Error())
		}
	}

	return &clusterv1.GetClusterResponse{Cluster: model.ConvertClusterToProto(&cluster)}, nil
}

func (a *api) UpdateCluster(ctx context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.UpdateClusterResponse, error) {
	var cluster model.Cluster

	fetchResult := a.gormDB.WithContext(ctx).First(&cluster, "id = ?", req.Cluster.GetId())
	if e := fetchResult.Error; e != nil {
		if errors.Is(e, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "cluster not found")
		} else {
			return nil, status.Error(codes.Internal, fetchResult.Error.Error())
		}
	}
	cluster.Name = req.Cluster.GetName()
	saveResult := a.gormDB.Save(&cluster)
	if e := saveResult.Error; e != nil {
		return nil, status.Error(codes.Internal, fetchResult.Error.Error())
	}

	return &clusterv1.UpdateClusterResponse{Cluster: model.ConvertClusterToProto(&cluster)}, nil
}

func (a *api) DeleteCluster(ctx context.Context, req *clusterv1.DeleteClusterRequest) (*clusterv1.DeleteClusterResponse, error) {
	result := a.gormDB.WithContext(ctx).Delete(&model.Cluster{}, "id = ?", req.GetId())
	if e := result.Error; e != nil {
		if errors.Is(e, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "cluster not found")
		} else {
			return nil, status.Error(codes.Internal, result.Error.Error())
		}
	}

	return &clusterv1.DeleteClusterResponse{}, nil
}
