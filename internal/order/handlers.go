package order

import (
	"encoding/json"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"go-otel-2/internal/models"
)

var tracer = otel.Tracer("order-service")

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "create_order")
	defer span.End()

	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if req.CustomerID == "" || req.Amount <= 0 {
		http.Error(w, `{"error":"customer_id and amount (positive) required"}`, http.StatusBadRequest)
		return
	}

	order := models.NewOrder(req.CustomerID, req.Amount)
	if err := h.repo.Create(ctx, order); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, `{"error":"failed to create order"}`, http.StatusInternalServerError)
		return
	}

	if h.metrics != nil {
		h.metrics.OrdersCreated.Add(ctx, 1)
	}

	span.SetAttributes(
		attribute.String("order.id", order.ID),
		attribute.String("order.customer_id", order.CustomerID),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(order)
}

