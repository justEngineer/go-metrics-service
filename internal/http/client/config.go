package client

import (
	"flag"
	"log"
	"os"
	"strconv"
)

type ClientConfig struct {
	endpoint       string
	reportInterval uint64
	pollInterval   uint64
	LogLevel       string
	SHA256Key      string
}

func Parse() ClientConfig {
	var cfg ClientConfig
	flag.StringVar(&cfg.endpoint, "a", "localhost:8080", "server host/port")
	flag.Uint64Var(&cfg.reportInterval, "r", 10, "update notification sending interval")
	flag.Uint64Var(&cfg.pollInterval, "p", 2, "polling stats interval")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.SHA256Key, "k", "", "SHA256 key")
	flag.Parse()
	if res := os.Getenv("ADDRESS"); res != "" {
		cfg.endpoint = res
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}
	if reportIntervalEnv := os.Getenv("REPORT_INTERVAL"); reportIntervalEnv != "" {
		res, err := strconv.Atoi(reportIntervalEnv)
		if err != nil {
			log.Fatal(err)
		}
		cfg.reportInterval = uint64(res)
	}
	if pollIntervalEnv := os.Getenv("POLL_INTERVAL"); pollIntervalEnv != "" {
		res, err := strconv.Atoi(pollIntervalEnv)
		if err != nil {
			log.Fatal(err)
		}
		cfg.pollInterval = uint64(res)
	}
	if res := os.Getenv("KEY"); res != "" {
		cfg.SHA256Key = res
	}
	return cfg
}
