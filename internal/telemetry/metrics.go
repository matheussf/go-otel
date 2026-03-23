package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// Metrics holds the application metrics.
type Metrics struct {
	OrdersCreated   metric.Int64Counter
	RequestDuration metric.Float64Histogram
}

// NewMetrics creates and registers application metrics.
func NewMetrics(ctx context.Context, meterName string) (*Metrics, error) {
	meter := otel.Meter(meterName)

	ordersCreated, err := meter.Int64Counter(
		"orders_created_total",
		metric.WithDescription("Total number of orders created"),
	)
	if err != nil {
		return nil, err
	}

	requestDuration, err := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		OrdersCreated:   ordersCreated,
		RequestDuration: requestDuration,
	}, nil
}
