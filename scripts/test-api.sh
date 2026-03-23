#!/bin/bash
# Test script for Order System API
# Run after: docker compose up -d

echo "Creating order..."
RESP=$(curl -s -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"cust-1","amount":99.99}')
echo "Response: $RESP"

if command -v jq &>/dev/null; then
  ORDER_ID=$(echo "$RESP" | jq -r '.id')
  if [ -n "$ORDER_ID" ] && [ "$ORDER_ID" != "null" ]; then
    echo "Fetching order $ORDER_ID..."
    curl -s "http://localhost:8080/orders/$ORDER_ID"
    echo ""
  fi
fi

echo "Done! Check Jaeger at http://localhost:16686"
