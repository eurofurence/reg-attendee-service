package ctlutil

import (
	"context"
	"encoding/json"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
	"net/http"
)

// --- response helpers ---

// WriteJson will also finalize the request, so if you don't want to return status 200, call WriteHeader first.
func WriteJson(ctx context.Context, w http.ResponseWriter, v interface{}) {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("error while encoding json response: %s", err.Error())
	}
}

func WriteHeader(ctx context.Context, w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	ctxvalues.SetHttpStatus(ctx, status)
}
