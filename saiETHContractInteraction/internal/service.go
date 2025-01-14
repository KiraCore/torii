package internal

import (
	"fmt"
	"sync"

	"github.com/saiset-co/sai-eth-interaction/models"
	saiService "github.com/saiset-co/sai-service/service"
	"go.uber.org/zap"
)

type InternalService struct {
	Handler   saiService.Handler // handlers to define in this specified microservice
	Contracts []models.Contract
	Mutex     *sync.RWMutex
	Context   *saiService.Context
	Logger    *zap.Logger
}

// global handler for registering handlers
var Service = &InternalService{
	Handler:   saiService.Handler{},
	Contracts: make([]models.Contract, 0),
	Mutex:     new(sync.RWMutex),
}

func (is *InternalService) Init() {
	fmt.Println(is.Context)
	Service.Logger = is.Context.Context.Value("logger").(*zap.Logger)

	Service.getInitialContracts("contracts.json")
}
