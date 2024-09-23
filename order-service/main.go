package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"order-service/handler"
	"order-service/service"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://mongodb:27017"))
	if err != nil {
		logger.Error("Failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	defer client.Disconnect(context.Background())

	// Initialize order service
	orderService := service.NewOrderService(client.Database("orders"), "amqp://guest:guest@rabbitmq:5672/")
	if orderService == nil {
		logger.Error("Failed to initialize OrderService")
		os.Exit(1)
	}

	// Initialize order handler
	orderHandler := handler.NewOrderHandler(orderService)

	// Set up HTTP routes
	http.HandleFunc("/order", orderHandler.CreateOrder)

	logger.Info("Starting order service on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
