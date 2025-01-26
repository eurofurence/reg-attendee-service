package ctxvalues

import (
	"context"
	"github.com/eurofurence/reg-attendee-service/docs"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValueMapOnEmptyContext(t *testing.T) {
	docs.Description("should return default value on uninitialized default context")
	require.Equal(t, "default", valueOrDefault(context.TODO(), "somekey", "default"), "unexpected value for uninitialized context")
}

func TestRetrieveRequestId(t *testing.T) {
	docs.Description("it should be possible to store and retrieve a requestId in an initialized context")
	ctx := CreateContextWithValueMap(context.TODO())
	SetRequestId(ctx, "hallo")
	require.Equal(t, "hallo", RequestId(ctx), "unexpected value retrieving request id that was just set")
}

func TestAsyncContextFrom(t *testing.T) {
	docs.Description("if present, the context values should carry over to an async context, but later changes should not")
	ctx := CreateContextWithValueMap(context.TODO())
	SetRequestId(ctx, "hallo")

	asyncCtx := AsyncContextFrom(ctx)
	SetRequestId(ctx, "changed")
	require.Equal(t, "hallo", RequestId(asyncCtx), "unexpected value retrieving request id")
}
