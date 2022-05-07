package ctlutil

import (
	"context"
	"encoding/json"
	"github.com/eurofurence/reg-attendee-service/internal/repository/logging"
	"github.com/eurofurence/reg-attendee-service/web/filter/ctxvalues"
	"net/http"
)

// --- response helpers ---

// WriteJson will also finalize the request, so if you don't want to return status 200, call WriteHeader first.
func WriteJson(ctx context.Context, w http.ResponseWriter, v interface{}) {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	if err != nil {
		logging.Ctx(ctx).Warnf("error while encoding json response: %v", err)
	}
}

func WriteHeader(ctx context.Context, w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	ctxvalues.SetHttpStatus(ctx, status)
}
