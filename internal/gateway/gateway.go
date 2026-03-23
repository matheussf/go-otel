package gateway

import (
	"io"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

var tracer = otel.Tracer("api-gateway")

// Proxy proxies requests to the order service with trace context propagation.
type Proxy struct {
	orderServiceURL string
	client          *http.Client
	propagator      propagation.TextMapPropagator
}

func NewProxy(orderServiceURL string) *Proxy {
	return &Proxy{
		orderServiceURL: orderServiceURL,
		client: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
		propagator: otel.GetTextMapPropagator(),
	}
}

func (p *Proxy) ForwardToOrderService(w http.ResponseWriter, r *http.Request, path, method string, body io.Reader) {
	ctx, span := tracer.Start(r.Context(), "gateway.forward_order_service")
	defer span.End()

	targetURL := p.orderServiceURL + path
	req, err := http.NewRequestWithContext(ctx, method, targetURL, body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	// Propagate trace context to downstream
	p.propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	// Copy relevant headers
	if r.Header.Get("Content-Type") != "" {
		req.Header.Set("Content-Type", r.Header.Get("Content-Type"))
	}

	resp, err := p.client.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, `{"error":"order service unreachable"}`, http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
