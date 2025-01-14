package app

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/saiset-co/saiEthIndexer/config"
	"github.com/saiset-co/saiEthIndexer/handlers"
	"github.com/saiset-co/saiEthIndexer/internal/http"
	"github.com/saiset-co/saiEthIndexer/pkg/httpserver"
	"github.com/saiset-co/saiEthIndexer/tasks"
	"go.uber.org/zap"
)

type App struct {
	Cfg         *config.Configuration
	Logger      *zap.Logger
	handlers    *handlers.Handlers
	taskManager *tasks.TaskManager
}

func New(args []string) (*App, error) {
	var (
		logger *zap.Logger
		err    error
	)
	if len(args) <= 1 || args[1] != "--debug" {
		logger, err = zap.NewProduction(zap.AddStacktrace(zap.DPanicLevel))
	} else {
		logger, err = zap.NewDevelopment(zap.AddStacktrace(zap.DPanicLevel))
	}
	if err != nil {
		return nil, err
	}

	logger.Debug("running logger in debug mode")
	if err != nil {
		log.Fatalf("error when start logger : %s", err)
	}
	return &App{
		Logger: logger,
	}, nil
}

// Register config to app
func (a *App) RegisterConfig(path string, contractsPath string) error {
	cfg := config.Configuration{}

	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("config read error: %w", err)
	}
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return fmt.Errorf("config unmarshal error: %w", err)
	}

	contracts := config.EthContracts{}
	b, err = os.ReadFile(contractsPath)
	if err != nil {
		return fmt.Errorf("contracts json read error: %w", err)
	}
	err = json.Unmarshal(b, &contracts)
	if err != nil {
		return fmt.Errorf("contracts json unmarshal error: %w", err)
	}
	cfg.EthContracts = contracts
	cfg.EthContracts.Mutex = new(sync.RWMutex)
	a.Cfg = &cfg
	a.Logger.Sugar().Debugf("start config :%+v\n", a.Cfg) // debug
	return nil
}

// Register task to app (main business logic)
func (a *App) RegisterTask(task *tasks.TaskManager) {
	a.taskManager = task
	return
}

// Register handlers to app
func (a *App) RegisterHandlers() {
	multihandler := handlers.Handlers{}
	if a.Cfg.Common.HttpServer.Enabled {
		//http server
		handler := gin.New()
		http.NewRouter(handler, a.Logger, a.taskManager)
		multihandler.Http = handler

	}

	a.handlers = &multihandler
}

func (a *App) Run() error {
	errChan := make(chan error, 1)
	var (
		httpServer = &httpserver.Server{}
	)
	if a.Cfg.Common.HttpServer.Enabled {
		httpServer = httpserver.New(a.handlers.Http, a.Cfg)
	}

	go a.taskManager.ProcessBlocks()

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		a.Logger.Error("app - Run - signal: " + s.String())
	case err := <-errChan:
		a.Logger.Error("app - Run - server notifier: ", zap.Error(err))
	}
	if a.Cfg.Common.HttpServer.Enabled {
		err := httpServer.Shutdown()
		if err != nil {
			a.Logger.Error("app - Run - httpServer.Shutdown: ", zap.Error(err))
			return err
		}
		a.Logger.Info("http server shutdown")
	}

	return nil
}
