# Order System MVP (Docker)

Sistema de pedidos mínimo com API Gateway e Order Service, instrumentado com OpenTelemetry.

## Arquitetura

```
Cliente → API Gateway (8080) → Order Service (8081) → PostgreSQL
                ↓                      ↓
            OTel Collector (4317) → Jaeger (16686)
```

## Requisitos

- Docker e Docker Compose
- Go 1.22+ (para build local)

## Executando

```bash
docker compose up --build
```

## Endpoints

- **POST /orders** – Criar pedido  
  ```json
  {"customer_id": "cust-123", "amount": 99.99}
  ```

- **GET /orders/{id}** – Buscar pedido por ID

## Testando

```bash
# Criar pedido
curl -X POST http://localhost:8080/orders -H "Content-Type: application/json" -d '{"customer_id":"cust-1","amount":99.99}'

# Buscar pedido (use o id retornado acima)
curl http://localhost:8080/orders/{order-id}
```

## Observabilidade

- **Jaeger UI**: http://localhost:16686
- Traces são propagados do Gateway até o Order Service
- OTel Collector recebe OTLP na porta 4317 (gRPC)

## Estrutura do Projeto

```
cmd/
  api-gateway/   – Gateway HTTP
  order-service/ – Serviço de pedidos
internal/
  gateway/       – Proxy e handlers do gateway
  order/         – Lógica de negócio e persistência
  models/        – Entidades compartilhadas
  telemetry/     – Setup OpenTelemetry
```

## Variáveis de Ambiente

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| PORT | 8080 / 8081 | Porta HTTP |
| ORDER_SERVICE_URL | http://order-service:8081 | URL do Order Service (gateway) |
| DATABASE_URL | postgres://... | Conexão PostgreSQL |
| OTEL_EXPORTER_OTLP_ENDPOINT | otel-collector:4317 | Endpoint OTLP para traces/metrics |
