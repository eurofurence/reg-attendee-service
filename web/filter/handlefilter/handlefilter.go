package handlefilter

import (
	"context"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/filter"
	"net/http"
)

type HandleFilter struct {
	handler filter.ContextAwareHandler
}

func Create(handler filter.ContextAwareHandler) filter.Filter {
	return &HandleFilter{handler: handler}
}

func (f *HandleFilter) Handle(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	f.handler(ctx, w, r)
}
