package model

import (
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	
	"go.admiral.io/admiral/api/cluster/v1"
)

type Cluster struct {
	Id   uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	Name string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (c *Cluster) BeforeCreate(_ *gorm.DB) (err error) {
	c.Id = uuid.New()
	return
}

func ConvertClusterToProto(c *Cluster) *clusterv1.Cluster {
	return &clusterv1.Cluster{
		Id:   c.Id.String(),
		Name: c.Name,

		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}
}
