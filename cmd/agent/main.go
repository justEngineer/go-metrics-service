package main

import (
	"time"

	storage "github.com/justEngineer/go-metrics-service/internal"
	client "github.com/justEngineer/go-metrics-service/internal/http/client"
)

func main() {
	MetricStorage := storage.New()
	config := client.Parse()
	ClientHandler := client.New(MetricStorage, &config)
	for {
		ClientHandler.OnTimer()
		time.Sleep(time.Second)
	}
}
