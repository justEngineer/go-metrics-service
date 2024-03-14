package main

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
}

func GetClientConfig() ClientConfig {
	var (
		endpoint       string
		reportInterval uint64
		pollInterval   uint64
	)
	flag.StringVar(&endpoint, "a", "localhost:8080", "server host/port")
	flag.Uint64Var(&reportInterval, "r", 10, "update notification sending interval")
	flag.Uint64Var(&pollInterval, "p", 2, "polling stats interval")
	flag.Parse()
	if res := os.Getenv("ADDRESS"); res != "" {
		endpoint = res
	}
	if reportIntervalEnv := os.Getenv("REPORT_INTERVAL"); reportIntervalEnv != "" {
		res, err := strconv.Atoi(reportIntervalEnv)

		if err != nil {
			log.Fatal(err)
		}

		reportInterval = uint64(res)
	}
	if pollIntervalEnv := os.Getenv("POLL_INTERVAL"); pollIntervalEnv != "" {
		res, err := strconv.Atoi(pollIntervalEnv)
		if err != nil {
			log.Fatal(err)
		}
		pollInterval = uint64(res)
	}
	return ClientConfig{
		endpoint,
		reportInterval,
		pollInterval,
	}
}
