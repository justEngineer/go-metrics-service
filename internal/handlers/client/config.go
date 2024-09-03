package client

import (
	"crypto/rsa"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/justEngineer/go-metrics-service/internal/encryption"
)

type ClientConfig struct {
	Endpoint        string `json:"address"`
	ReportInterval  uint64 `json:"poll_interval"`
	PollInterval    uint64 `json:"report_interval"`
	LogLevel        string
	SHA256Key       string
	RateLimit       uint64
	PublicKeyPath   string `json:"crypto_key"`
	PublicCryptoKey *rsa.PublicKey
}

func loadConfigFromFile(path string) (ClientConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return ClientConfig{}, err
	}
	defer file.Close()

	var config ClientConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return ClientConfig{}, err
	}
	if config.PublicKeyPath != "" {
		var err error
		config.PublicCryptoKey, err = encryption.GetPublicKey(config.PublicKeyPath)
		if err != nil {
			log.Fatalf("RSA private key read error:%s", err)
		}
	}

	return config, nil
}

func Parse() ClientConfig {
	var cfg ClientConfig
	var publicKeyPath string
	var configFilePath string
	flag.StringVar(&cfg.Endpoint, "a", "localhost:8080", "server host/port")
	flag.Uint64Var(&cfg.ReportInterval, "r", 10, "update notification sending interval")
	flag.Uint64Var(&cfg.PollInterval, "p", 2, "polling stats interval")
	flag.StringVar(&cfg.LogLevel, "lg", "info", "log level")
	flag.StringVar(&cfg.SHA256Key, "k", "", "SHA256 key")
	flag.StringVar(&publicKeyPath, "crypto-key", "", "path to the public encryption key")
	flag.Uint64Var(&cfg.RateLimit, "l", 1, "max rate limit of outgoing requests")
	flag.StringVar(&configFilePath, "c", "", "path to the configuration file")
	flag.Parse()
	if res := os.Getenv("ADDRESS"); res != "" {
		cfg.Endpoint = res
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}
	if reportIntervalEnv := os.Getenv("REPORT_INTERVAL"); reportIntervalEnv != "" {
		res, err := strconv.Atoi(reportIntervalEnv)
		if err != nil {
			log.Fatal(err)
		}
		cfg.ReportInterval = uint64(res)
	}
	if pollIntervalEnv := os.Getenv("POLL_INTERVAL"); pollIntervalEnv != "" {
		res, err := strconv.Atoi(pollIntervalEnv)
		if err != nil {
			log.Fatal(err)
		}
		cfg.PollInterval = uint64(res)
	}
	if res := os.Getenv("KEY"); res != "" {
		cfg.SHA256Key = res
	}
	if res := os.Getenv("RATE_LIMIT"); res != "" {
		value, err := strconv.Atoi(res)
		if err != nil {
			log.Fatal(err)
		}
		cfg.RateLimit = uint64(value)
	}
	if cryptoKeyEnv := os.Getenv("CRYPTO_KEY"); cryptoKeyEnv != "" {
		publicKeyPath = cryptoKeyEnv
	}
	if publicKeyPath != "" {
		var err error
		cfg.PublicCryptoKey, err = encryption.GetPublicKey(publicKeyPath)
		if err != nil {
			log.Fatalf("RSA public key read error:%s", err)
		}
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
		if cfg.ReportInterval == 0 {
			cfg.ReportInterval = fileConfig.ReportInterval
		}
		if cfg.PollInterval == 0 {
			cfg.PollInterval = fileConfig.PollInterval
		}
		if cfg.PublicCryptoKey.Size() == 0 {
			cfg.PublicCryptoKey = fileConfig.PublicCryptoKey
		}
	}
	return cfg
}
