package main

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"order-service/model"
	"order-service/service"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func BenchmarkCreateOrder(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		logger.Error("Failed to connect to MongoDB", "error", err)
		b.FailNow()
	}
	defer client.Disconnect(context.Background())

	// Initialize order service
	orderService := service.NewOrderService(client.Database("orders"), "amqp://guest:guest@localhost:5672/")
	if orderService == nil {
		logger.Error("Failed to initialize OrderService")
		b.FailNow()
	}

	// Benchmark the CreateOrder method
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		order := &model.Order{
			UserID:   "user123",
			Product:  "product123",
			Quantity: 1,
		}
		_, err := orderService.CreateOrder(context.Background(), order)
		if err != nil {
			logger.Error("Failed to create order", "error", err)
			b.FailNow()
		}
	}
}

func TestMain(m *testing.M) {
	// Setup code if needed
	code := m.Run()
	// Teardown code if needed
	os.Exit(code)
}
