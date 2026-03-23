package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go-otel-2/internal/order"
	"go-otel-2/internal/telemetry"
)

const (
	defaultPort        = "8081"
	defaultDBConn      = "postgres://postgres:postgres@postgres:5432/orders?sslmode=disable"
	defaultOTLPEndpoint = "otel-collector:4317"
)

func main() {
	port := getEnv("PORT", defaultPort)
	dbConn := getEnv("DATABASE_URL", defaultDBConn)
	otlpEndpoint := getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", defaultOTLPEndpoint)

	ctx := context.Background()

	// Init telemetry
	tp, err := telemetry.InitTracer(ctx, "order-service", otlpEndpoint)
	if err != nil {
		log.Printf("Warn: failed to init tracer: %v (continuing without tracing)", err)
	} else {
		defer func() {
			_ = tp.Shutdown(ctx)
		}()
	}

	mp, err := telemetry.InitMeterProvider(ctx, "order-service", otlpEndpoint)
	if err != nil {
		log.Printf("Warn: failed to init meter provider: %v (continuing without metrics)", err)
	} else {
		defer func() {
			_ = mp.Shutdown(ctx)
		}()
	}

	metrics, err := telemetry.NewMetrics(ctx, "order-service")
	if err != nil {
		log.Printf("Warn: failed to init metrics: %v", err)
	}

	// Init DB
	db, err := sql.Open("pgx", dbConn)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	if err := initDB(ctx, db); err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}

	repo := order.NewRepository(db)
	orderHandler := order.NewHandler(repo, metrics)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /orders", orderHandler.CreateOrder)
	mux.HandleFunc("GET /orders/{id}", orderHandler.GetOrderHandler)

	handler := otelhttp.NewHandler(mux, "order-service", otelhttp.WithServerName("order-service"))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	go func() {
		log.Printf("Order service listening on :%s", port)
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

func initDB(ctx context.Context, db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS orders (
		id VARCHAR(36) PRIMARY KEY,
		customer_id VARCHAR(255) NOT NULL,
		amount DECIMAL(12,2) NOT NULL,
		status VARCHAR(50) NOT NULL,
		created_at TIMESTAMP NOT NULL
	);
	`
	_, err := db.ExecContext(ctx, query)
	return err
}
