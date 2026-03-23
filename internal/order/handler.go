package order

import (
	"encoding/json"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"go-otel-2/internal/telemetry"
)

type Handler struct {
	repo    *Repository
	metrics *telemetry.Metrics
}

func NewHandler(repo *Repository, metrics *telemetry.Metrics) *Handler {
	return &Handler{repo: repo, metrics: metrics}
}

// getOrderID extracts order ID from path like /orders/{id}
func getOrderID(path string) string {
	const prefix = "/orders/"
	if strings.HasPrefix(path, prefix) {
		return strings.TrimPrefix(path, prefix)
	}
	return ""
}

func (h *Handler) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	id := getOrderID(r.URL.Path)
	if id == "" {
		http.Error(w, `{"error":"order id required"}`, http.StatusBadRequest)
		return
	}

	h.GetOrderByID(w, r, id)
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		id = getOrderID(r.URL.Path)
	}
	if id == "" {
		http.Error(w, `{"error":"order id required"}`, http.StatusBadRequest)
		return
	}

	h.GetOrderByID(w, r, id)
}

func (h *Handler) GetOrderByID(w http.ResponseWriter, r *http.Request, id string) {
	tracer := otel.Tracer("order-service")

	ctx, span := tracer.Start(r.Context(), "get_order")
	defer span.End()

	span.SetAttributes(attribute.String("order.id", id))

	order, err := h.repo.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		http.Error(w, `{"error":"failed to get order"}`, http.StatusInternalServerError)
		return
	}

	if order == nil {
		http.Error(w, `{"error":"order not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(order)
}