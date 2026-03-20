package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/linkai/go-chatbot-api/adapters/system"
	domainerrors "github.com/linkai/go-chatbot-api/domain/errors"
	"github.com/linkai/go-chatbot-api/domain/model"
	"github.com/linkai/go-chatbot-api/domain/ports"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("httpapi")

type SessionResolver interface {
	ResolveAuthenticatedProfile(context.Context, ports.AuthenticatedUser) (model.UserProfile, error)
}

func RequestIDMiddleware(ctx huma.Context, next func(huma.Context)) {
	requestID := strings.TrimSpace(ctx.Header("X-Request-ID"))
	if requestID == "" {
		requestID = system.NewIDGenerator().NewID("req")
	}
	ctx.SetHeader("X-Request-ID", requestID)
	ctx = huma.WithValue(ctx, requestIDContextKey, requestID)
	next(ctx)
}

func OtelTracingMiddleware(ctx huma.Context, next func(huma.Context)) {
	request, _ := humago.Unwrap(ctx)
	spanCtx, span := tracer.Start(
		request.Context(),
		fmt.Sprintf("%s %s", request.Method, request.URL.Path),
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", request.Method),
		attribute.String("http.route", request.URL.Path),
	)

	next(huma.WithValue(ctx, "otel_ctx", spanCtx))

	status := ctx.Status()
	span.SetAttributes(attribute.Int("http.status_code", status))
	if status >= http.StatusInternalServerError {
		span.SetStatus(codes.Error, "server error")
	}
}

func AuthMiddleware(api huma.API, authenticator ports.SessionAuthenticator, sessions SessionResolver) func(huma.Context, func(huma.Context)) {
	if authenticator == nil {
		panic("AuthMiddleware: authenticator is required")
	}
	if sessions == nil {
		panic("AuthMiddleware: sessions is required")
	}
	return func(ctx huma.Context, next func(huma.Context)) {
		authHeader := strings.TrimSpace(ctx.Header("Authorization"))
		if authHeader == "" {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "missing authorization header")
			return
		}
		authUser, err := authenticator.Authenticate(ctx.Context(), authHeader)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, err.Error())
			return
		}
		profile, err := sessions.ResolveAuthenticatedProfile(ctx.Context(), authUser)
		if err != nil {
			writeMappedError(api, ctx, err)
			return
		}
		next(huma.WithValue(ctx, userProfileContextKey, profile))
	}
}

func AdminOnlyMiddleware(api huma.API) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		profile, ok := UserProfileFromContext(ctx.Context())
		if !ok || profile.Role != model.RoleAdmin {
			huma.WriteErr(api, ctx, http.StatusForbidden, "admin access required")
			return
		}
		next(ctx)
	}
}

func writeMappedError(api huma.API, ctx huma.Context, err error) {
	status := http.StatusInternalServerError
	var validationErr *domainerrors.ValidationError
	switch {
	case errors.Is(err, domainerrors.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, domainerrors.ErrInvalidInput), errors.As(err, &validationErr):
		status = http.StatusUnprocessableEntity
	case errors.Is(err, domainerrors.ErrAlreadyExists), errors.Is(err, domainerrors.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, domainerrors.ErrUnauthorized):
		status = http.StatusUnauthorized
	case errors.Is(err, domainerrors.ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, domainerrors.ErrTokenBudgetExceeded):
		status = http.StatusTooManyRequests
	}
	huma.WriteErr(api, ctx, status, err.Error())
}
