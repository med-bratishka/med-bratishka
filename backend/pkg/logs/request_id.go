package logs

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const RequestIDContextKey contextKey = "request_id"

func CtxWithRequestID(ctx context.Context) context.Context {
	return context.WithValue(ctx, RequestIDContextKey, uuid.NewString())
}

func RequestIDFromContext(ctx context.Context) string {
	rid, _ := ctx.Value(RequestIDContextKey).(string)
	return rid
}
