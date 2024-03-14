package main

import (
	"flag"
	"os"
)

type ServerConfig struct {
	endpoint string
}

func GetServerConfig() ServerConfig {
	var serverHostPort string
	flag.StringVar(&serverHostPort, "a", "localhost:8080", "server host/port")
	flag.Parse()
	if res := os.Getenv("ADDRESS"); res != "" {
		serverHostPort = res
	}
	return ServerConfig{
		serverHostPort,
	}
}
