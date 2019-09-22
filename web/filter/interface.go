package filter

import (
	"context"
	"net/http"
)

type ContextAwareHandler = func(ctx context.Context, w http.ResponseWriter, r *http.Request)

type Filter interface {
	Handle(ctx context.Context, w http.ResponseWriter, r *http.Request)
}
