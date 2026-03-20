package model

import (
	"strings"
	"time"

	domainerrors "github.com/linkai/go-chatbot-api/domain/errors"
)

type ConversationStatus string

const (
	ConversationStatusActive   ConversationStatus = "active"
	ConversationStatusArchived ConversationStatus = "archived"
)

type ChatMessageRole string

const (
	ChatMessageRoleUser      ChatMessageRole = "user"
	ChatMessageRoleAssistant ChatMessageRole = "assistant"
)

type Conversation struct {
	ID         string
	UserID     string
	ChatbotID  string
	Status     ConversationStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ArchivedAt *time.Time
}

type Message struct {
	ID             string
	ConversationID string
	Role           ChatMessageRole
	Content        string
	Sequence       int
	InputTokens    int
	OutputTokens   int
	CreatedAt      time.Time
}

func NewConversation(id, userID, chatbotID string, now time.Time) (Conversation, error) {
	conversation := Conversation{
		ID:        strings.TrimSpace(id),
		UserID:    strings.TrimSpace(userID),
		ChatbotID: strings.TrimSpace(chatbotID),
		Status:    ConversationStatusActive,
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC(),
	}
	if err := conversation.Validate(); err != nil {
		return Conversation{}, err
	}
	return conversation, nil
}

func (c Conversation) Validate() error {
	if c.ID == "" {
		return &domainerrors.ValidationError{Field: "id", Message: "is required"}
	}
	if c.UserID == "" {
		return &domainerrors.ValidationError{Field: "user_id", Message: "is required"}
	}
	if c.ChatbotID == "" {
		return &domainerrors.ValidationError{Field: "chatbot_id", Message: "is required"}
	}
	switch c.Status {
	case ConversationStatusActive, ConversationStatusArchived:
	default:
		return &domainerrors.ValidationError{Field: "status", Message: "is invalid"}
	}
	return nil
}

func (c Conversation) Archive(now time.Time) Conversation {
	archivedAt := now.UTC()
	c.Status = ConversationStatusArchived
	c.ArchivedAt = &archivedAt
	c.UpdatedAt = archivedAt
	return c
}

func NewMessage(id, conversationID string, role ChatMessageRole, content string, sequence int, now time.Time) (Message, error) {
	message := Message{
		ID:             strings.TrimSpace(id),
		ConversationID: strings.TrimSpace(conversationID),
		Role:           role,
		Content:        strings.TrimSpace(content),
		Sequence:       sequence,
		CreatedAt:      now.UTC(),
	}
	if err := message.Validate(); err != nil {
		return Message{}, err
	}
	return message, nil
}

func (m Message) Validate() error {
	if m.ID == "" {
		return &domainerrors.ValidationError{Field: "id", Message: "is required"}
	}
	if m.ConversationID == "" {
		return &domainerrors.ValidationError{Field: "conversation_id", Message: "is required"}
	}
	switch m.Role {
	case ChatMessageRoleUser, ChatMessageRoleAssistant:
	default:
		return &domainerrors.ValidationError{Field: "role", Message: "is invalid"}
	}
	if m.Content == "" {
		return &domainerrors.ValidationError{Field: "content", Message: "is required"}
	}
	if m.Sequence <= 0 {
		return &domainerrors.ValidationError{Field: "sequence", Message: "must be positive"}
	}
	return nil
}
