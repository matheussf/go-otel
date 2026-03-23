package gateway

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var gatewayTracer = otel.Tracer("api-gateway")

func (p *Proxy) HandleCreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx, span := gatewayTracer.Start(r.Context(), "gateway.create_order")
	defer span.End()

	span.SetAttributes(attribute.String("http.method", r.Method), attribute.String("http.route", "/orders"))

	p.ForwardToOrderService(w, r.WithContext(ctx), "/orders", http.MethodPost, r.Body)
}

func (p *Proxy) HandleGetOrder(w http.ResponseWriter, r *http.Request) {
	ctx, span := gatewayTracer.Start(r.Context(), "gateway.get_order")
	defer span.End()

	id := r.PathValue("id")
	if id == "" {
		id = getOrderIDFromPath(r.URL.Path)
	}
	if id == "" {
		http.Error(w, `{"error":"order id required"}`, http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.route", "/orders/{id}"),
		attribute.String("order.id", id),
	)

	p.ForwardToOrderService(w, r.WithContext(ctx), "/orders/"+id, http.MethodGet, nil)
}

func getOrderIDFromPath(path string) string {
	const prefix = "/orders/"
	if len(path) > len(prefix) && path[:len(prefix)] == prefix {
		return path[len(prefix):]
	}
	return ""
}
