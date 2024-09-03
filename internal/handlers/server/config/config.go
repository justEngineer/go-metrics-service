package config

import (
	"crypto/rsa"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/justEngineer/go-metrics-service/internal/encryption"
)

// ServerConfig содержит конфигурацию для сервера.
type ServerConfig struct {
	Endpoint         string          `json:"address"` // URL-адрес конечной точки сервера
	LogLevel         string          // Уровень логирования
	StoreInterval    int             `json:"store_interval"` // Интервал сохранения данных
	FileStorePath    string          `json:"store_file"`     // Путь к файлу с архивом хранения данных
	Restore          bool            `json:"restore"`        // Флаг восстановления данных из архива
	DatabaseDSN      string          `json:"database_dsn"`   // Строка подключения к базе данных
	SHA256Key        string          // Ключ для подписи данных
	PrivateCryptoKey *rsa.PrivateKey // Ключ для шифрования данных
	PrivateKeyPath   string          `json:"crypto_key"`     // Путь к файлу Ключ для шифрования данных
	TrustedSubnet    string          `yaml:"trusted_subnet"` //	Cтроковое представление бесклассовой адресации (CIDR)
}

func loadConfigFromFile(path string) (ServerConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return ServerConfig{}, err
	}
	defer file.Close()

	var config ServerConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return ServerConfig{}, err
	}
	if config.PrivateKeyPath != "" {
		var err error
		config.PrivateCryptoKey, err = encryption.GetPrivateKey(config.PrivateKeyPath)
		if err != nil {
			log.Fatalf("RSA private key read error:%s", err)
		}
	}
	return config, nil
}

// Parse функция чтения конфигурации
func Parse() ServerConfig {
	var cfg ServerConfig
	var privateKeyPath string
	var configFilePath string
	flag.StringVar(&cfg.Endpoint, "a", "localhost:8080", "server host/port")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "store interval file")
	flag.StringVar(&cfg.FileStorePath, "f", "/tmp/metrics-db.json", "path file")
	flag.BoolVar(&cfg.Restore, "r", true, "restore file")
	//flag.StringVar(&cfg.DatabaseDSN, "d", "postgres://postgres:admin@localhost:5432/postgres?sslmode=disable", "postgres database connection string")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "postgres database connection string")
	flag.StringVar(&cfg.SHA256Key, "k", "", "SHA256 key")
	flag.StringVar(&privateKeyPath, "crypto-key", "", "path to the private encryption key")
	flag.StringVar(&configFilePath, "c", "", "path to the configuration file")
	flag.StringVar(&cfg.TrustedSubnet, "t", "", "Trusted subnet (CIDR)")
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
		cfg.DatabaseDSN = res
	}
	if res := os.Getenv("KEY"); res != "" {
		cfg.SHA256Key = res
	}
	if cryptoKeyEnv := os.Getenv("CRYPTO_KEY"); cryptoKeyEnv != "" {
		privateKeyPath = cryptoKeyEnv
	}
	if privateKeyPath != "" {
		var err error
		cfg.PrivateCryptoKey, err = encryption.GetPrivateKey(privateKeyPath)
		if err != nil {
			log.Fatalf("RSA private key read error:%s", err)
		}
	}
	if res := os.Getenv("TRUSTED_SUBNET"); res != "" {
		cfg.TrustedSubnet = res
	}

	if configFileEnv := os.Getenv("CONFIG"); configFileEnv != "" {
		configFilePath = configFileEnv
	}
	if configFilePath != "" {
		fileConfig, err := loadConfigFromFile(configFilePath)
		if err != nil {
			log.Fatalf("Error loading config file: %v, path: %s", err, configFilePath)
		}
		if cfg.Endpoint == "" {
			cfg.Endpoint = fileConfig.Endpoint
		}
		if cfg.StoreInterval == 0 {
			cfg.StoreInterval = fileConfig.StoreInterval
		}
		if cfg.FileStorePath == "" {
			cfg.FileStorePath = fileConfig.FileStorePath
		}
		if !cfg.Restore {
			cfg.Restore = fileConfig.Restore
		}
		if cfg.DatabaseDSN == "" {
			cfg.DatabaseDSN = fileConfig.DatabaseDSN
		}
		if cfg.PrivateKeyPath == "" {
			cfg.PrivateKeyPath = fileConfig.PrivateKeyPath
		}
	}

	return cfg
}
