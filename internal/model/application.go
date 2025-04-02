package model

import (
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	
	"go.admiral.io/admiral/api/application/v1"
)

type Application struct {
	Id   uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	Name string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (a *Application) BeforeCreate(_ *gorm.DB) (err error) {
	a.Id = uuid.New()
	return
}

func ConvertApplicationToProto(a *Application) *applicationv1.Application {
	return &applicationv1.Application{
		Id:   a.Id.String(),
		Name: a.Name,

		CreatedAt: timestamppb.New(a.CreatedAt),
		UpdatedAt: timestamppb.New(a.UpdatedAt),
	}
}
