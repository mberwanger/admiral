package application

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uber-go/tally/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	applicationv1 "go.admiral.io/admiral/api/application/v1"
	"go.admiral.io/admiral/internal/config"
	"go.admiral.io/admiral/internal/endpoint"
	"go.admiral.io/admiral/internal/model"
	"go.admiral.io/admiral/internal/service"
	"go.admiral.io/admiral/internal/service/database"
)

const Name = "endpoint.tenant"

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
	applicationv1.RegisterApplicationAPIServer(r.GRPCServer(), a)
	return r.RegisterJSONGateway(applicationv1.RegisterApplicationAPIHandler)
}

func (a *api) CreateApplication(ctx context.Context, req *applicationv1.CreateApplicationRequest) (*applicationv1.CreateApplicationResponse, error) {
	application := model.Application{
		Name: req.GetName(),
	}

	result := a.gormDB.WithContext(ctx).Create(&application)
	if result.Error != nil {
		return nil, result.Error
	}

	return &applicationv1.CreateApplicationResponse{Application: model.ConvertApplicationToProto(&application)}, nil
}

func (a *api) ListApplications(ctx context.Context, req *applicationv1.ListApplicationsRequest) (*applicationv1.ListApplicationsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (a *api) GetApplication(ctx context.Context, req *applicationv1.GetApplicationRequest) (*applicationv1.GetApplicationResponse, error) {
	var application model.Application

	result := a.gormDB.WithContext(ctx).First(&application, "id = ?", req.GetId())
	if e := result.Error; e != nil {
		if errors.Is(e, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "tenant not found")
		} else {
			return nil, status.Error(codes.Internal, result.Error.Error())
		}
	}

	return &applicationv1.GetApplicationResponse{Application: model.ConvertApplicationToProto(&application)}, nil
}

func (a *api) UpdateApplication(ctx context.Context, req *applicationv1.UpdateApplicationRequest) (*applicationv1.UpdateApplicationResponse, error) {
	var application model.Application

	fetchResult := a.gormDB.WithContext(ctx).First(&application, "id = ?", req.Application.GetId())
	if e := fetchResult.Error; e != nil {
		if errors.Is(e, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "application not found")
		} else {
			return nil, status.Error(codes.Internal, fetchResult.Error.Error())
		}
	}
	application.Name = req.Application.GetName()
	saveResult := a.gormDB.Save(&application)
	if e := saveResult.Error; e != nil {
		return nil, status.Error(codes.Internal, fetchResult.Error.Error())
	}

	return &applicationv1.UpdateApplicationResponse{Application: model.ConvertApplicationToProto(&application)}, nil
}

func (a *api) DeleteApplication(ctx context.Context, req *applicationv1.DeleteApplicationRequest) (*applicationv1.DeleteApplicationResponse, error) {
	result := a.gormDB.WithContext(ctx).Delete(&model.Application{}, "id = ?", req.GetId())
	if e := result.Error; e != nil {
		if errors.Is(e, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "application not found")
		} else {
			return nil, status.Error(codes.Internal, result.Error.Error())
		}
	}

	return &applicationv1.DeleteApplicationResponse{}, nil
}
