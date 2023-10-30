package internal

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/KiraCore/sekai-bridge/logger"
	"github.com/KiraCore/sekai-bridge/tss"

	"github.com/saiset-co/saiP2P-go/config"
	p2p "github.com/saiset-co/saiP2P-go/core"
	"github.com/saiset-co/saiService"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type InternalService struct {
	Context *saiService.Context
	P2P     *p2p.Core
	Tss     *tss.TssServer
}

func (is *InternalService) Init() {
	// logger
	l := logger.Init(is.Context.GetConfig("common.log_mode", "debug").(string))

	//@TODO: change config, need partyID in it
	conf, err := config.Get()
	if err != nil {
		logger.Logger.Fatal("config.Get", zap.Error(err))
	}

	tssConf, err := GetConfig()
	if err != nil {
		logger.Logger.Fatal("GetConfig", zap.Error(err))
	}

	// p2p initialization
	testFilterFunc := func(interface{}) bool {
		return true
	}

	is.P2P = p2p.Init(conf, testFilterFunc)
	fmt.Printf("p2p conf %+v", is.P2P.Config)

	go is.P2P.Run(testFilterFunc)

	// tss initialization
	tssServer := tss.New(tssConf.Tss.Pubkey, is.P2P, l)
	is.Tss = tssServer

	// start keygen instance
	tssKeygen := is.Tss.NewTssKeyGen()

	is.Tss.TssKeygen = tssKeygen

	go func() {
		for {
			p2pMsg, err := is.P2P.NextMsg(context.Background())
			if err != nil {
				is.P2P.Logger.Error("internal -> service -> Init -> NextMsg", zap.Error(err))
				continue
			}

			err = is.Tss.HandleP2Pmessage(p2pMsg)
			if err != nil {
				is.P2P.Logger.Error("internal -> service -> Init -> HandleP2Pmessage", zap.Error(err))
				continue
			}
		}
	}()

	// @TODO:something instead of time.Sleep
	time.Sleep(5 * time.Second)

	// send tss handshake to add values to map[partyID]peerAddr
	if len(is.P2P.ConnectionStorage) > 0 {
		for addr := range is.P2P.ConnectionStorage {
			err := is.Tss.SendHandshake(addr)
			if err != nil {
				is.P2P.Logger.Error("internal -> service -> Init -> Handshake", zap.String("addr", addr), zap.Error(err))
				continue
			}
		}
	}

	//go is.Tss.EventListening()

	l.Debug("is.Init", zap.Any("tss", is.Tss))

}

func (is InternalService) Process() {
}

// Tss config model
type Config struct {
	P2P struct {
		Port  string   `yaml:"port"`
		Slot  int      `yaml:"slot"`
		Peers []string `yaml:"peers"`
	} `yaml:"p2p"`
	Http struct {
		Enabled bool   `yaml:"enabled"`
		Port    string `yaml:"port"`
	} `yaml:"http"`
	Tss struct {
		Pubkey string `yaml:"pubkey"`
	} `yaml:"tss"`

	OnBroadcastMessageReceive []string
	OnDirectMessageReceive    []string
	DebugMode                 bool `yaml:"debug"`
}

// Get - parses config.yml, return config struct
func GetConfig() (Config, error) {
	config := Config{}
	yamlData, err := os.ReadFile("config.yml")

	if err != nil {
		return config, fmt.Errorf("Readfile : %w", err)
	}

	err = yaml.Unmarshal(yamlData, &config)

	if err != nil {
		return config, fmt.Errorf("Unmarshal : %w", err)
	}
	return config, nil
}
