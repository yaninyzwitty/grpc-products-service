package main

import (
	"log/slog"
	"os"

	"github.com/yaninyzwitty/grpc-products-service/internal/pkg"
)

func main() {
	var cfg pkg.Config
	file, err := os.Open("config.yaml") //TODO- change this
	if err != nil {
		slog.Error("failed to open config.yaml", "error", err)
		os.Exit(1)
	}
	defer file.Close()

	if err := cfg.LoadConfig(file); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

}
