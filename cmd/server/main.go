package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	storage "github.com/justEngineer/go-metrics-service/internal"
	server "github.com/justEngineer/go-metrics-service/internal/http/server"
)

func main() {
	config := server.Parse()
	MetricStorage := storage.New()
	ServerHandler := server.New(MetricStorage, &config)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Post("/update/{type}/{name}/{value}", ServerHandler.UpdateMetric)
	r.Get("/value/{type}/{name}", ServerHandler.GetMetric)
	r.Get("/", ServerHandler.MainPage)

	port := strings.Split(config.Endpoint, ":")
	log.Fatal(http.ListenAndServe(":"+port[1], r))
}
