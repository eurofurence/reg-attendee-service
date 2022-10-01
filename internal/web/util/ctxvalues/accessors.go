package ctxvalues

import (
	"context"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
)

const ContextMap = "map"

const ContextRequestId = "requestid"
const ContextBearerToken = "bearertoken"
const ContextApiToken = "apitoken"
const ContextAuthorizedAs = "authorizedas"
const ContextEmail = "email"
const ContextName = "name"
const ContextSubject = "subject"

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

func Email(ctx context.Context) string {
	return valueOrDefault(ctx, ContextEmail, "")
}

func SetEmail(ctx context.Context, email string) {
	setValue(ctx, ContextEmail, email)
}

func Name(ctx context.Context) string {
	return valueOrDefault(ctx, ContextName, "")
}

func SetName(ctx context.Context, Name string) {
	setValue(ctx, ContextName, Name)
}

func Subject(ctx context.Context) string {
	return valueOrDefault(ctx, ContextSubject, "")
}

func SetSubject(ctx context.Context, Subject string) {
	setValue(ctx, ContextSubject, Subject)
}

func HasApiToken(ctx context.Context) bool {
	v := valueOrDefault(ctx, ContextApiToken, "")
	return v == config.FixedApiToken()
}

func SetApiToken(ctx context.Context, apiToken string) {
	setValue(ctx, ContextApiToken, apiToken)
}

func IsAuthorizedAsRole(ctx context.Context, role string) bool {
	value := valueOrDefault(ctx, fmt.Sprintf("%s-%s", ContextAuthorizedAs, role), "")
	return value == role
}

func SetAuthorizedAsRole(ctx context.Context, role string) {
	setValue(ctx, fmt.Sprintf("%s-%s", ContextAuthorizedAs, role), role)
}
