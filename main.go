package main

import (
	"github.com/KiraCore/sekai-bridge/internal"
	"github.com/saiset-co/saiService"
)

var id int

func main() {

	svc := saiService.NewService("sekai-bridge")
	is := internal.InternalService{Context: svc.Context}

	svc.RegisterConfig("config.yml")

	svc.RegisterInitTask(is.Init)

	svc.RegisterTasks([]func(){
		is.Process,
	})

	svc.RegisterHandlers(
		is.NewHandler(),
	)
	svc.Start()
}
