package ports

import (
	"context"
	"io"
	"time"

	"github.com/linkai/go-chatbot-api/domain/model"
)

type IdentityUser struct {
	ID    string
	Email string
}

type AuthenticatedUser struct {
	ID    string
	Email string
}

type ExtractedDocument struct {
	Text string
	Kind model.KnowledgeFileKind
}

type AIChatMessage struct {
	Role    model.ChatMessageRole
	Content string
}

type AIChatRequest struct {
	Model        string
	SystemPrompt string
	Conversation []AIChatMessage
}

type AIChatResponse struct {
	Content      string
	InputTokens  int
	OutputTokens int
	FinishReason string
}

type IdentityAdmin interface {
	InviteUser(context.Context, string, string) (IdentityUser, error)
	DisableUser(context.Context, string) error
	DeleteUser(context.Context, string) error
	CreateUser(context.Context, string, string, bool) (IdentityUser, error)
}

type SessionAuthenticator interface {
	Authenticate(context.Context, string) (AuthenticatedUser, error)
}

type ObjectStorage interface {
	PutObject(context.Context, string, string, io.Reader, int64) error
	DeleteObject(context.Context, string) error
}

type TextExtractor interface {
	Extract(context.Context, string, string, []byte) (ExtractedDocument, error)
}

type TokenCounter interface {
	CountText(string) int
}

type AIChatClient interface {
	Complete(context.Context, AIChatRequest) (AIChatResponse, error)
}

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	NewID(string) string
}

type Logger interface {
	Debug(context.Context, string, ...any)
	Info(context.Context, string, ...any)
	Warn(context.Context, string, ...any)
	Error(context.Context, string, ...any)
}
