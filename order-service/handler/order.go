package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"order-service/model"
	"order-service/service"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var order model.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		slog.Error("Failed to decode order", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	createdOrder, err := h.orderService.CreateOrder(r.Context(), &order)
	if err != nil {
		slog.Error("Failed to create order", "error", err)
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdOrder)
}
