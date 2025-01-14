package internal

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KiraCore/sekai-bridge/logger"
	"github.com/KiraCore/sekai-bridge/tss"
	"github.com/KiraCore/sekai-bridge/utils"
	"github.com/gorilla/mux"

	"github.com/saiset-co/saiP2P-go/config"
	p2p "github.com/saiset-co/saiP2P-go/core"
	"github.com/saiset-co/saiService"
	"go.uber.org/zap"
)

type InternalService struct {
	Context *saiService.Context
	P2P     *p2p.Core
	Tss     *tss.TssServer
	Logger  *zap.Logger
}

func (is *InternalService) Init() {
	// logger
	is.Logger = logger.Init(is.Context.GetConfig("common.log_mode", "debug").(string))

	// @TODO: change config, need partyID in it
	conf, err := config.Get()
	if err != nil {
		is.Logger.Fatal("config.Get", zap.Error(err))
	}

	tssConf, err := utils.GetConfig()
	if err != nil {
		is.Logger.Fatal("GetConfig", zap.Error(err))
	}

	// p2p initialization
	testFilterFunc := func(interface{}) bool {
		return true
	}

	is.P2P = p2p.Init(conf, testFilterFunc)

	go is.P2P.Run(testFilterFunc)

	// tss initialization
	tssServer := tss.New(tssConf.Tss.PublicKey, tssConf.Tss.Parties,
		tssConf.Tss.Threshold, tssConf.Tss.Quorum, is.P2P, is.Logger)

	is.Tss = tssServer

	key, err := utils.LoadKeyFile()
	if err == nil {
		is.Tss.Key = key
		is.Tss.Logger.Info("key loaded", zap.String("pub", key.ECDSAPub.Y().String()),
			zap.String("pub base64 encoded", base64.StdEncoding.EncodeToString(key.ECDSAPub.Bytes())))
	} else {
		is.Tss.Logger.Info("key was not found")
	}

	go func() {
		for {
			p2pMsg, err := is.P2P.NextMsg(context.Background())
			if err != nil {
				is.P2P.Logger.Error("internal -> service -> Init -> NextMsg", zap.Error(err))
				continue
			}

			if p2pMsg == nil {
				continue
			}

			go is.Tss.HandleP2Pmessage(p2pMsg)
		}
	}()

	time.Sleep(1 * time.Second)

	if len(is.P2P.ConnectionStorage) > 0 {
		for addr := range is.P2P.ConnectionStorage {
			err := is.Tss.SendHandshake(addr)
			if err != nil {
				is.P2P.Logger.Error("internal -> service -> Init -> Handshake", zap.String("addr", addr), zap.Error(err))
				continue
			}
		}
	}

	// graceful shutdown
	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

		for {
			s := <-interrupt
			var (
				errCh    = make(chan error)
				resultCh = make(chan bool)
			)
			is.P2P.Logger.Info("internal -> service -> got interrupt", zap.String("signal", s.String()))
			is.P2P.Disconnect()                       // p2p notifying
			go is.Tss.SendDisconnect(errCh, resultCh) // tss notifying
			select {
			case <-time.After(5 * time.Second): // @TODO: config value?
				is.P2P.Logger.Info("timeout expired, exiting app.... ")
				os.Exit(0)
			case err := <-errCh:
				is.P2P.Logger.Error("service -> SendDisconnect", zap.Error(err))
				is.P2P.Logger.Info("exiting app...")
				os.Exit(0)
			case result := <-resultCh:
				is.P2P.Logger.Sugar().Infof("graceful shutdown result = %t,exiting app...", result)
				os.Exit(0)
			}
		}
	}()

	go func() {
		router := mux.NewRouter()
		router.Handle("/", http.FileServer(http.Dir("./html")))

		if err := http.ListenAndServe(":"+conf.Http.Port, router); err != nil {
			log.Println("Http server error: ", err)
		}
	}()

}

func (is *InternalService) Process() {
}
