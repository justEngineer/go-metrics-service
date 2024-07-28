package config

import (
	"flag"
	"log"
	"os"
	"strconv"
)

// ServerConfig содержит конфигурацию для сервера.
type ServerConfig struct {
	Endpoint      string // URL-адрес конечной точки сервера
	LogLevel      string // Уровень логирования
	StoreInterval int    // Интервал сохранения данных
	FileStorePath string // Пуь к файлу с архивом хранения данных
	Restore       bool   // Флаг восстановления данных из архива
	DBConnection  string // Строка подключения к базе данных
	SHA256Key     string // Ключ для подписи данных
}

// Parse функция чтения конфигурации
func Parse() ServerConfig {
	var cfg ServerConfig
	flag.StringVar(&cfg.Endpoint, "a", "localhost:8080", "server host/port")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "store interval file")
	flag.StringVar(&cfg.FileStorePath, "f", "/tmp/metrics-db.json", "path file")
	flag.BoolVar(&cfg.Restore, "r", true, "restore file")
	//flag.StringVar(&cfg.DBConnection, "d", "postgres://postgres:admin@localhost:5432/postgres?sslmode=disable", "postgres database connection string")
	flag.StringVar(&cfg.DBConnection, "d", "", "postgres database connection string")
	flag.StringVar(&cfg.SHA256Key, "k", "", "SHA256 key")
	flag.Parse()
	if res := os.Getenv("ADDRESS"); res != "" {
		cfg.Endpoint = res
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}
	if res := os.Getenv("STORE_INTERVAL"); res != "" {
		value, err := strconv.Atoi(res)
		if err != nil || value < 0 {
			log.Println("STORE_INTERVAL argument parse failed", err)
		} else {
			cfg.StoreInterval = value
		}
	}
	if res := os.Getenv("FILE_STORAGE_PATH"); res != "" {
		cfg.FileStorePath = res
	}
	if res := os.Getenv("RESTORE"); res != "" {
		value, err := strconv.ParseBool(res)
		if err != nil {
			log.Println("RESTORE argument parse failed", err)
		} else {
			cfg.Restore = value
		}
	}
	if res := os.Getenv("DATABASE_DSN"); res != "" {
		cfg.DBConnection = res
	}
	if res := os.Getenv("KEY"); res != "" {
		cfg.SHA256Key = res
	}
	return cfg
}
