package client

import (
	"crypto/rsa"
	"flag"
	"log"
	"os"
	"strconv"

	security "github.com/justEngineer/go-metrics-service/internal/security"
)

type ClientConfig struct {
	endpoint        string
	reportInterval  uint64
	pollInterval    uint64
	LogLevel        string
	SHA256Key       string
	RateLimit       uint64
	PublicCryptoKey *rsa.PublicKey // Ключ для шифрования данных
}

func Parse() ClientConfig {
	var cfg ClientConfig
	var publicKeyPath string
	flag.StringVar(&cfg.endpoint, "a", "localhost:8080", "server host/port")
	flag.Uint64Var(&cfg.reportInterval, "r", 10, "update notification sending interval")
	flag.Uint64Var(&cfg.pollInterval, "p", 2, "polling stats interval")
	flag.StringVar(&cfg.LogLevel, "lg", "info", "log level")
	flag.StringVar(&cfg.SHA256Key, "k", "", "SHA256 key")
	flag.StringVar(&publicKeyPath, "crypto-key", "", "path to the public encryption key")
	flag.Uint64Var(&cfg.RateLimit, "l", 1, "max rate limit of outgoing requests")
	flag.Parse()
	if res := os.Getenv("ADDRESS"); res != "" {
		cfg.endpoint = res
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}
	if reportIntervalEnv := os.Getenv("REPORT_INTERVAL"); reportIntervalEnv != "" {
		res, err := strconv.Atoi(reportIntervalEnv)
		if err != nil {
			log.Fatal(err)
		}
		cfg.reportInterval = uint64(res)
	}
	if pollIntervalEnv := os.Getenv("POLL_INTERVAL"); pollIntervalEnv != "" {
		res, err := strconv.Atoi(pollIntervalEnv)
		if err != nil {
			log.Fatal(err)
		}
		cfg.pollInterval = uint64(res)
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
		cfg.PublicCryptoKey, err = security.GetPublicKey(publicKeyPath)
		if err != nil {
			log.Fatalf("RSA private key read error:%s", err)
		}
	}
	return cfg
}
