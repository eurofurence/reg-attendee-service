package middleware

import (
	"net/http"

	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Tracing wraps the handler with OpenTelemetry HTTP instrumentation.
// It automatically creates spans for incoming requests and propagates trace context.
// When telemetry is disabled, the global TracerProvider is a no-op, so this middleware
// has negligible overhead.
func Tracing(next http.Handler) http.Handler {
	return otelhttp.NewHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add request ID as span attribute if available
			if reqID := ctxvalues.RequestId(r.Context()); reqID != "" {
				span := trace.SpanFromContext(r.Context())
				span.SetAttributes(attribute.String("request.id", reqID))
			}
			next.ServeHTTP(w, r)
		}),
		"http-request",
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			// Use chi's route pattern for low-cardinality span names
			// e.g., "GET /api/attendees/{id}" instead of "GET /api/attendees/123"
			if routePattern := chi.RouteContext(r.Context()).RoutePattern(); routePattern != "" {
				return r.Method + " " + routePattern
			}
			return r.Method + " " + r.URL.Path
		}),
	)
}
