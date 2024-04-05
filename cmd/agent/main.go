package main

import (
	"context"
	"os/signal"
	"sync"
	"syscall"

	client "github.com/justEngineer/go-metrics-service/internal/http/client"
	storage "github.com/justEngineer/go-metrics-service/internal/storage"
)

func main() {
	MetricStorage := storage.New()
	config := client.Parse()
	ClientHandler := client.New(MetricStorage, &config)

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	//defer
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		ClientHandler.GetMetrics(ctx)
		stop()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		ClientHandler.SendMetrics(ctx)
		stop()
	}()
	wg.Wait()
}
