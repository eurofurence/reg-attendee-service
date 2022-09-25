package ctxvalues

import (
	"context"
	"errors"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"net/http"
	"strconv"
)

const ContextMap = "map"

const ContextHttpStatusKey = "httpstatus"
const ContextRequestId = "requestid"
const ContextBearerToken = "bearertoken"
const ContextAuthorizedAs = "authorizedas"

func CreateContextWithValueMap(ctx context.Context) context.Context {
	// this is so we can add values to our context, like ... I don't know ... the http status from the response!
	contextMap := make(map[string]string)

	ctx = context.WithValue(ctx, ContextMap, contextMap)
	return ctx
}

func valueOrDefault(ctx context.Context, key string, defaultValue string) string {
	contextMapUntyped := ctx.Value(ContextMap)
	if contextMapUntyped == nil {
		return defaultValue
	}
	contextMap := contextMapUntyped.(map[string]string)

	if val, ok := contextMap[key]; ok {
		return val
	} else {
		return defaultValue
	}
}

func setValue(ctx context.Context, key string, value string) {
	contextMapUntyped := ctx.Value(ContextMap)
	if contextMapUntyped != nil {
		contextMap := contextMapUntyped.(map[string]string)
		contextMap[key] = value
	}
}

func HttpStatus(ctx context.Context) string {
	return valueOrDefault(ctx, ContextHttpStatusKey, fmt.Sprint(http.StatusOK))
}

func SetHttpStatus(ctx context.Context, status int) {
	setValue(ctx, ContextHttpStatusKey, fmt.Sprint(status))
}

func RequestId(ctx context.Context) string {
	return valueOrDefault(ctx, ContextRequestId, "00000000")
}

func SetRequestId(ctx context.Context, requestId string) {
	setValue(ctx, ContextRequestId, requestId)
}

func BearerToken(ctx context.Context) string {
	return valueOrDefault(ctx, ContextBearerToken, "")
}

func SetBearerToken(ctx context.Context, bearerToken string) {
	setValue(ctx, ContextBearerToken, bearerToken)
}

func AuthorizedAsGroup(ctx context.Context) (config.FixedTokenEnum, error) {
	authStr := valueOrDefault(ctx, ContextAuthorizedAs, "")
	if authStr == "" {
		return -1, errors.New("no authorization entry found")
	} else {
		i, err := strconv.Atoi(authStr)
		if err != nil {
			return -1, err
		}
		return config.FixedTokenEnum(i), nil
	}
}

func SetAuthorizedAsGroup(ctx context.Context, group config.FixedTokenEnum) {
	setValue(ctx, ContextAuthorizedAs, fmt.Sprint(group))
}
