package main

import (
	"encoding/json"
	"fmt"

	"github.com/saiset-co/sai-service/service"
	"github.com/saiset-co/sai-storage-mongo/internal"
	"github.com/saiset-co/sai-storage-mongo/logger"
	"github.com/saiset-co/sai-storage-mongo/mongo"
	"github.com/saiset-co/sai-storage-mongo/types"
)

func main() {
	name := "SaiStorage"

	svc := service.NewService(name)

	svc.RegisterConfig("config.yml")

	logger.Logger = svc.Logger

	config, err := convertConfig(svc.GetConfig("storage", nil))
	if err != nil {
		fmt.Println("Could not read configuration:", err)
	}

	client, err := mongo.NewMongoClient(config)
	if err != nil {
		fmt.Println("Could not connect to the mongo server:", err)
	}

	defer client.Host.Disconnect(svc.Context.Context)

	is := internal.InternalService{
		Name:    name,
		Context: svc.Context,
		Client:  client,
	}

	svc.RegisterTasks([]func(){client.DuplicateProcessor})

	svc.RegisterHandlers(
		is.NewHandler(),
	)

	svc.Start()
}

func convertConfig(data interface{}) (*types.StorageConfig, error) {
	var config = new(types.StorageConfig)

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(jsonBytes, config)
	if err != nil {
		return config, err
	}

	return config, nil
}
