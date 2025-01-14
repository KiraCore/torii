package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config model
type Config struct {
	P2P struct {
		Port string `yaml:"port"`
		Slot int    `yaml:"slot"`
		Http struct {
			Port string
		}
	} `yaml:"p2p"`
	Http struct {
		Port string
	}
	Peers                     []string `yaml:"peers"`
	OnBroadcastMessageReceive []string
	OnDirectMessageReceive    []string
	DebugMode                 bool `yaml:"debug"`
	UDP                       struct {
		MsgBufferSize int           `yaml:"bufferSize"` // buffer for udp msg 49900 is max in ubuntu
		Interval      time.Duration `yaml:"interval"`   // interval for p2p msg sending when
	}
	Cache struct { // cache settings
		TTL         int `yaml:"ttl"`   // ttl for entry in cache
		CleanPeriod int `yaml:"clean"` // when ttl expired entries should be cleaned
	} `yaml:"cache"`
}

// Get - parses config.yml, return config struct
func Get() (Config, error) {
	config := Config{}
	yamlData, err := os.ReadFile("config.yml")

	if err != nil {
		return config, fmt.Errorf("Readfile : %w", err)
	}

	err = yaml.Unmarshal(yamlData, &config)

	if err != nil {
		return config, fmt.Errorf("Unmarshal : %w", err)
	}

	fmt.Printf("p2pConf : %+v", config)
	return config, nil

}
