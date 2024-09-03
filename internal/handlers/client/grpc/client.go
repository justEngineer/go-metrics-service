// Package proto предназначен для хранения абстракций, связанных с gRPC.
package grpc

import (
	"context"
	"log"

	model "github.com/justEngineer/go-metrics-service/internal/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/justEngineer/go-metrics-service/internal/proto"
)

func SendMetricModelData(ctx context.Context, data []model.Metrics) error {
	address := ":9090"
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewMetricsClient(conn)
	batch := make([]*pb.Metric, len(data))
	for i, metric := range data {
		if metric.Value == nil {
			batch[i] = &pb.Metric{
				MType: metric.MType,
				ID:    metric.ID,
				Delta: metric.Delta,
			}
		} else {
			batch[i] = &pb.Metric{
				MType: metric.MType,
				ID:    metric.ID,
				Value: metric.Value,
			}
		}
	}

	_, err = client.UpdateMetrics(ctx, &pb.UpdateMetricsRequest{
		Metrics: batch,
	})
	return err
}
