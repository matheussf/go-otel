package models

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	Amount     float64   `json:"amount"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreateOrderRequest struct {
	CustomerID string  `json:"customer_id"`
	Amount     float64 `json:"amount"`
}

func NewOrder(customerID string, amount float64) *Order {
	return &Order{
		ID:         uuid.New().String(),
		CustomerID: customerID,
		Amount:     amount,
		Status:     "pending",
		CreatedAt:  time.Now(),
	}
}
