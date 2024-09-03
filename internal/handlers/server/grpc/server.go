// Package proto предназначен для хранения абстракций, связанных с gRPC.
package grpc

import (
	"context"
	"log"
	"net"

	logger "github.com/justEngineer/go-metrics-service/internal/logger"
	pb "github.com/justEngineer/go-metrics-service/internal/proto"
	"github.com/justEngineer/go-metrics-service/internal/storage"
	"google.golang.org/grpc"
)

type Server struct {
	storage   storage.Storage
	appLogger *logger.Logger
	pb.UnimplementedMetricsServer
}

// New создает новый экземпляр Handler.
func NewMetricsServer(storage storage.Storage, log *logger.Logger) *Server {
	return &Server{storage: storage, appLogger: log}
}

// UpdateMetrics is a stream method updater for metrics
func (s *Server) UpdateMetrics(ctx context.Context, in *pb.UpdateMetricsRequest) (*pb.UpdateMetricResponse, error) {
	var gaugeMetrics []storage.GaugeMetric
	var counterMetrics []storage.CounterMetric

	for _, parameter := range in.Metrics {
		switch parameter.MType {
		case "gauge":
			gaugeMetrics = append(gaugeMetrics, storage.GaugeMetric{Name: parameter.ID, Value: *parameter.Value})
		case "counter":
			counterMetrics = append(counterMetrics, storage.CounterMetric{Name: parameter.ID, Value: *parameter.Delta})
		default:
			s.appLogger.Log.Warn("Unkniwn metrict type")
		}
	}
	err := s.storage.SetMetricsBatch(ctx, gaugeMetrics, counterMetrics)
	return nil, err
}

func (s *Server) Start(ctx context.Context) *grpc.Server {
	listen, err := net.Listen("tcp", "localhost:9090")
	if err != nil {
		log.Fatalf("gRPC server start failed, error: %v", err)
	}
	grpcServer := grpc.NewServer()
	go func() {
		pb.RegisterMetricsServer(grpcServer, s)
		if err := grpcServer.Serve(listen); err != nil {
			log.Fatalf("gRPC server start failed, error: %v", err)
		}
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()
	return grpcServer
}
