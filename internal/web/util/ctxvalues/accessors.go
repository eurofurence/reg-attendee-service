package ctxvalues

import (
	"context"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"strings"
)

const ContextMap = "map"

const ContextRequestId = "requestid"
const ContextIdToken = "idtoken"
const ContextAccessToken = "accesstoken"
const ContextApiToken = "apitoken"
const ContextAuthorizedAs = "authorizedas"
const ContextEmail = "email"
const ContextEmailVerified = "emailverified"
const ContextName = "name"
const ContextSubject = "subject"
const ContextAvatar = "avatar"

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

func IdToken(ctx context.Context) string {
	return valueOrDefault(ctx, ContextIdToken, "")
}

func SetIdToken(ctx context.Context, token string) {
	setValue(ctx, ContextIdToken, token)
}

func AccessToken(ctx context.Context) string {
	return valueOrDefault(ctx, ContextAccessToken, "")
}

func SetAccessToken(ctx context.Context, token string) {
	setValue(ctx, ContextAccessToken, token)
}

func Email(ctx context.Context) string {
	return valueOrDefault(ctx, ContextEmail, "")
}

func SetEmail(ctx context.Context, email string) {
	setValue(ctx, ContextEmail, email)
}

func EmailVerified(ctx context.Context) bool {
	valueStr := valueOrDefault(ctx, ContextEmailVerified, "false")
	return valueStr == "true"
}

func SetEmailVerified(ctx context.Context, verified bool) {
	if verified {
		setValue(ctx, ContextEmailVerified, "true")
	}
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

func Avatar(ctx context.Context) string {
	return valueOrDefault(ctx, ContextAvatar, "")
}

func SetAvatar(ctx context.Context, avatar string) {
	setValue(ctx, ContextAvatar, avatar)
}

func HasApiToken(ctx context.Context) bool {
	v := valueOrDefault(ctx, ContextApiToken, "")
	return v == config.FixedApiToken()
}

func SetApiToken(ctx context.Context, apiToken string) {
	setValue(ctx, ContextApiToken, apiToken)
}

func IsAuthorizedAsGroup(ctx context.Context, group string) bool {
	value := valueOrDefault(ctx, fmt.Sprintf("%s-%s", ContextAuthorizedAs, group), "")
	return value == group
}

func SetAuthorizedAsGroup(ctx context.Context, group string) {
	setValue(ctx, fmt.Sprintf("%s-%s", ContextAuthorizedAs, group), group)
}

func ClearAuthorizedGroups(ctx context.Context) {
	contextMapUntyped := ctx.Value(ContextMap)
	if contextMapUntyped != nil {
		contextMap := contextMapUntyped.(map[string]string)
		for k, _ := range contextMap {
			if strings.HasPrefix(k, ContextAuthorizedAs) {
				delete(contextMap, k)
			}
		}
	}
}
