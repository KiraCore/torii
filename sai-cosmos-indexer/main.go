package main

import (
	"github.com/saiset-co/sai-service-crud-plus/logger"
	"github.com/saiset-co/saiCosmosIndexer/internal"
	"github.com/saiset-co/saiService"
)

func main() {
	svc := saiService.NewService("saiCosmosIndexer")
	is := internal.InternalService{Context: svc.Context}

	svc.RegisterConfig("config.yml")

	logger.Logger = svc.Logger

	svc.RegisterInitTask(is.Init)

	svc.RegisterTasks([]func(){
		is.Process,
	})

	svc.RegisterHandlers(
		is.NewHandler(),
	)

	svc.Start()
}
