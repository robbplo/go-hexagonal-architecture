package httpapi

import (
	"context"

	"github.com/linkai/go-chatbot-api/domain/model"
)

type contextKey string

const (
	requestIDContextKey   contextKey = "request_id"
	userProfileContextKey contextKey = "user_profile"
)

func withRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(requestIDContextKey).(string)
	return value
}

func withUserProfile(ctx context.Context, profile model.UserProfile) context.Context {
	return context.WithValue(ctx, userProfileContextKey, profile)
}

func UserProfileFromContext(ctx context.Context) (model.UserProfile, bool) {
	profile, ok := ctx.Value(userProfileContextKey).(model.UserProfile)
	return profile, ok
}
