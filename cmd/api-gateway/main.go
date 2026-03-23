package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go-otel-2/internal/gateway"
	"go-otel-2/internal/telemetry"
)

const (
	defaultPort          = "8080"
	defaultOrderService  = "http://order-service:8081"
	defaultOTLPEndpoint  = "otel-collector:4317"
)

func main() {
	port := getEnv("PORT", defaultPort)
	orderServiceURL := getEnv("ORDER_SERVICE_URL", defaultOrderService)
	otlpEndpoint := getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", defaultOTLPEndpoint)

	ctx := context.Background()

	// Init telemetry
	tp, err := telemetry.InitTracer(ctx, "api-gateway", otlpEndpoint)
	if err != nil {
		log.Printf("Warn: failed to init tracer: %v (continuing without tracing)", err)
	} else {
		defer func() {
			_ = tp.Shutdown(ctx)
		}()
	}

	_, _ = telemetry.InitMeterProvider(ctx, "api-gateway", otlpEndpoint)

	proxy := gateway.NewProxy(orderServiceURL)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /orders", proxy.HandleCreateOrder)
	mux.HandleFunc("GET /orders/{id}", proxy.HandleGetOrder)

	handler := otelhttp.NewHandler(mux, "api-gateway", otelhttp.WithServerName("api-gateway"))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	go func() {
		log.Printf("API Gateway listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
