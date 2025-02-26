package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yaninyzwitty/grpc-products-service/helpers"
	"github.com/yaninyzwitty/grpc-products-service/internal/controllers"
	"github.com/yaninyzwitty/grpc-products-service/internal/database"
	"github.com/yaninyzwitty/grpc-products-service/internal/queue"
	"github.com/yaninyzwitty/grpc-products-service/pb"
	"github.com/yaninyzwitty/grpc-products-service/snowflake"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := helpers.FetchFromAWSConfig(ctx)
	if err != nil {
		slog.Error("failed to load config from AWS", "error", err)
		os.Exit(1)
	}

	err = snowflake.InitSonyFlake()
	if err != nil {
		slog.Error("failed to initialize snowflake", "error", err)
		os.Exit(1)
	}

	astraCfg := &database.AstraConfig{
		Username: cfg.Database.Username,
		Path:     cfg.Database.Path,
		Token:    helpers.GetEnvOrDefault("ASTRA_TOKEN", ""),
	}

	db := database.NewAstraDB()

	session, err := db.Connect(ctx, astraCfg, 30*time.Second)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	defer session.Close()

	pulsarCfg := &queue.PulsarConfig{
		URI:       cfg.Queue.Uri,
		Token:     helpers.GetEnvOrDefault("PULSAR_TOKEN", ""),
		TopicName: cfg.Queue.TopicName,
	}

	pulsarClient, err := pulsarCfg.CreatePulsarConnection(ctx)
	if err != nil {
		slog.Error("failed to create pulsar connection", "error", err)
		os.Exit(1)
	}
	defer pulsarClient.Close()

	pulsarProducer, err := pulsarCfg.CreatePulsarProducer(ctx, pulsarClient)
	if err != nil {
		slog.Error("failed to create pulsar producer", "error", err)
		os.Exit(1)
	}

	defer pulsarProducer.Close()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		slog.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	productController := controllers.NewProductController(session)
	server := grpc.NewServer()
	reflection.Register(server)

	pb.RegisterProductsServiceServer(server, productController)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	stopCH := make(chan os.Signal, 1)

	// polling approach

	go func() {
		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// poll messages
				if err := helpers.ProcessMessages(context.Background(), session, pulsarProducer); err != nil {
					slog.Error("failed to process messages", "error", err)
					os.Exit(1)
				}
			case <-stopCH:
				return
			}

		}
	}()

	go func() {
		sig := <-sigChan
		slog.Info("Received shutdown signal", "signal", sig)
		slog.Info("Shutting down gRPC server...")

		// Gracefully stop the gRPC server
		server.GracefulStop()
		cancel() // Cancel context for other goroutines
		slog.Info("gRPC server has been stopped gracefully")
	}()

	slog.Info("Starting gRPC server", "port", cfg.Server.Port)
	if err := server.Serve(lis); err != nil {
		slog.Error("gRPC server encountered an error while serving", "error", err)
		os.Exit(1)
	}

}
