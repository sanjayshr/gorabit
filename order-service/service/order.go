package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"order-service/model"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderService struct {
	db         *mongo.Database
	collection *mongo.Collection
	rabbitMQ   *amqp.Connection
}

func NewOrderService(db *mongo.Database, rabbitMQURL string) *OrderService {
	var conn *amqp.Connection
	var err error
	for i := 0; i < 30; i++ { // Retry up to 30 times
		conn, err = amqp.Dial(rabbitMQURL)
		if err == nil {
			break
		}
		slog.Error("Failed to connect to RabbitMQ, retrying...", "error", err)
		time.Sleep(2 * time.Second) // Wait for 2 seconds before retrying
	}
	if err != nil {
		slog.Error("Failed to connect to RabbitMQ after retries", "error", err)
		return nil
	}
	return &OrderService{
		db:         db,
		collection: db.Collection("orders"),
		rabbitMQ:   conn,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, order *model.Order) (*model.Order, error) {
	order.ID = primitive.NewObjectID().Hex()

	_, err := s.collection.InsertOne(ctx, order)
	if err != nil {
		return nil, err
	}

	// Publish order created event
	err = s.publishOrderCreatedEvent(order)
	if err != nil {
		slog.Error("Failed to publish order created event", "error", err)
	}

	return order, nil
}

func (s *OrderService) publishOrderCreatedEvent(order *model.Order) error {
	ch, err := s.rabbitMQ.Channel()
	if err != nil {
		return err
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
		return err
	}

	body, err := json.Marshal(order)
	if err != nil {
		return err
	}

	return ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}
