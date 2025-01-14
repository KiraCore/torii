package internal

import (
	"github.com/saiset-co/sai-service/service"
	"github.com/saiset-co/sai-storage-mongo/mongo"
)

type InternalService struct {
	Name    string
	Context *service.Context
	Client  *mongo.Client
}
