package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/streadway/amqp"
)

type Order struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Product  string `json:"product"`
	Quantity int    `json:"quantity"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	var conn *amqp.Connection
	var err error
	for i := 0; i < 30; i++ { // Retry up to 30 times
		conn, err = amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
		if err == nil {
			break
		}
		logger.Error("Failed to connect to RabbitMQ, retrying...", "error", err)
		time.Sleep(2 * time.Second) // Wait for 2 seconds before retrying
	}
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ after retries", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		logger.Error("Failed to open a channel", "error", err)
		os.Exit(1)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"order_created", // queue name
		false,           // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		logger.Error("Failed to declare a queue", "error", err)
		os.Exit(1)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		logger.Error("Failed to register a consumer", "error", err)
		os.Exit(1)
	}

	logger.Info("Notification service is running")

	for d := range msgs {
		var order Order
		if err := json.Unmarshal(d.Body, &order); err != nil {
			logger.Error("Failed to unmarshal order", "error", err)
			continue
		}

		logger.Info("Received order", "order_id", order.ID, "user_id", order.UserID, "product", order.Product, "quantity", order.Quantity)
		// Here you would implement the actual notification logic (e.g., sending an email or SMS)
	}

	logger.Info("Starting notification service on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		logger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
