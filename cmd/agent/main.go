package main

import (
	"context"
	"os/signal"
	"sync"
	"syscall"

	storage "github.com/justEngineer/go-metrics-service/internal"
	client "github.com/justEngineer/go-metrics-service/internal/http/client"
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
