package ctlutil

import (
	"context"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

// --- parameter parsers ---

// note, if these return an error, you must remember to bail out

func AttendeeIdFromVars(ctx context.Context, w http.ResponseWriter, r *http.Request) (uint, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		InvalidAttendeeIdErrorHandler(ctx, w, r, idStr)
	}
	return uint(id), err
}
