package server

import (
	"flag"
	"os"
)

const DefaultPathToHTMLTemplate = "../internal/http/server/main_page_html.tmpl"

type ServerConfig struct {
	Endpoint           string
	PathToHTMLTemplate string
	LogLevel           string
}

func Parse() ServerConfig {
	var cfg ServerConfig
	flag.StringVar(&cfg.Endpoint, "a", "localhost:8080", "server host/port")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.PathToHTMLTemplate, "p", DefaultPathToHTMLTemplate,
		"relative path to the html template file")
	flag.Parse()
	if res := os.Getenv("ADDRESS"); res != "" {
		cfg.Endpoint = res
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}
	return cfg
}
