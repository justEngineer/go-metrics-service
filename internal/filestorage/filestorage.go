package filestorage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	server "github.com/justEngineer/go-metrics-service/internal/http/server"
	logger "github.com/justEngineer/go-metrics-service/internal/logger"
	storage "github.com/justEngineer/go-metrics-service/internal/storage"
	"go.uber.org/zap"
)

type FileStorage struct {
	storage *storage.MemStorage
	config  *server.ServerConfig
}

func (fs FileStorage) SaveDumpToFile() error {
	rawData := fs.storage.GetAllMetrics()
	jsonData, err := json.Marshal(rawData)
	if err != nil {
		return fmt.Errorf("serializing dump to JSON failed: %s", err)
	}
	err = os.WriteFile(fs.config.FileStorePath, jsonData, 0666)
	if err != nil {
		return fmt.Errorf("write to file failed: %s", err)
	}
	return nil
}

func New(metricStorage *storage.MemStorage, config *server.ServerConfig, ctx context.Context, logger *logger.Logger) *FileStorage {
	if config.FileStorePath == "" {
		return nil
	}
	fileStorage := &FileStorage{
		storage: metricStorage,
		config:  config,
	}
	if config.Restore {
		data, err := os.ReadFile(config.FileStorePath)
		if err == nil {
			var backupData storage.MetricsDump
			err = json.Unmarshal(data, &backupData)
			if err != nil {
				panic("cannot deserialize JSON data from file")
			}
			metricStorage.Mutex.Lock()
			for _, counter := range backupData.Counters {
				metricStorage.Counter[counter.Name] = counter.Value
			}
			for _, gauge := range backupData.Gauges {
				metricStorage.Gauge[gauge.Name] = gauge.Value
			}
			metricStorage.Mutex.Unlock()
		} else if !os.IsNotExist(err) {
			panic("cannot read dump, file doesn't exists")
		}
	}
	if config.StoreInterval > 0 {
		go func() {
			for {
				storeTicker := time.NewTicker(time.Duration(config.StoreInterval) * time.Second)
				defer storeTicker.Stop()

				for {
					select {
					case <-ctx.Done():
						return
					case <-storeTicker.C:
						if err := fileStorage.SaveDumpToFile(); err != nil {
							logger.Log.Error("error while saving dump to file", zap.Error(err))
							continue
						}
					}
				}
			}
		}()
	}
	return fileStorage
}
