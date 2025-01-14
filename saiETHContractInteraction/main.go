package main

import (
	"github.com/saiset-co/sai-eth-interaction/internal"
	saiService "github.com/saiset-co/sai-service/service"
)

func main() {
	svc := saiService.NewService("saiEthInteraction")

	svc.RegisterConfig("config.yml")

	is := internal.InternalService{Context: svc.Context}

	svc.RegisterInitTask(is.Init)

	svc.RegisterHandlers(
		is.NewHandler())

	svc.Start()

}
