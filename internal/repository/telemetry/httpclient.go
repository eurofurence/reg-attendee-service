package telemetry

import (
	"net/http"

	aurestclientapi "github.com/StephanHCB/go-autumn-restclient/api"
	auresthttpclient "github.com/StephanHCB/go-autumn-restclient/implementation/httpclient"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// NewTracingTransport returns an http.RoundTripper that wraps the default
// transport with OpenTelemetry instrumentation. This automatically:
// - Creates child spans for outgoing HTTP requests
// - Propagates trace context via W3C TraceContext headers
// - Records HTTP semantic conventions (method, status, url, etc.)
//
// When telemetry is disabled, the global TracerProvider is a no-op,
// so this has negligible overhead.
func NewTracingTransport() http.RoundTripper {
	return otelhttp.NewTransport(http.DefaultTransport)
}

// NewHttpClient creates an aurestclientapi.Client with OpenTelemetry tracing enabled.
// This wraps the HTTP transport with otelhttp to automatically create spans
// for outgoing requests and propagate trace context.
func NewHttpClient(requestManipulator aurestclientapi.RequestManipulatorCallback) aurestclientapi.Client {
	transport := NewTracingTransport()
	return &auresthttpclient.HttpClientImpl{
		HttpClient: &http.Client{
			Transport: transport,
		},
		RequestManipulator: requestManipulator,
	}
}
