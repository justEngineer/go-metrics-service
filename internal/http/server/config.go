package server

import (
	"flag"
	"log"
	"os"
	"strconv"
)

type ServerConfig struct {
	Endpoint      string
	LogLevel      string
	StoreInterval int
	FileStorePath string
	Restore       bool
}

func Parse() ServerConfig {
	var cfg ServerConfig
	flag.StringVar(&cfg.Endpoint, "a", "localhost:8080", "server host/port")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "store interval file")
	flag.StringVar(&cfg.FileStorePath, "f", "/tmp/metrics-db.json", "path file")
	flag.BoolVar(&cfg.Restore, "r", true, "restore file")
	flag.Parse()
	if res := os.Getenv("ADDRESS"); res != "" {
		cfg.Endpoint = res
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}
	if res := os.Getenv("STORE_INTERVAL"); res != "" {
		value, err := strconv.Atoi(res)
		if err != nil || value < 0 {
			log.Println("STORE_INTERVAL argument parse failed", err)
		} else {
			cfg.StoreInterval = value
		}
	}
	if res := os.Getenv("FILE_STORAGE_PATH"); res != "" {
		cfg.FileStorePath = res
	}
	if res := os.Getenv("RESTORE"); res != "" {
		value, err := strconv.ParseBool(res)
		if err != nil {
			log.Println("RESTORE argument parse failed", err)
		} else {
			cfg.Restore = value
		}
	}
	return cfg
}
