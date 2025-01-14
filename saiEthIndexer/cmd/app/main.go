package main

import (
	"log"
	"os"

	"github.com/saiset-co/saiEthIndexer/internal/app"
	"github.com/saiset-co/saiEthIndexer/tasks"
)

func main() {
	args := os.Args

	app, err := app.New(args)
	if err != nil {
		log.Fatal(err)
	}

	//register config with specific options
	err = app.RegisterConfig("./config.json", "./contracts.json")
	if err != nil {
		log.Fatal(err)
	}

	t, err := tasks.NewManager(app.Cfg, app.Logger)
	if err != nil {
		log.Fatal(err)
	}

	defer t.Logger.Sync()

	app.RegisterTask(t)

	app.RegisterHandlers()

	app.Run()

}
