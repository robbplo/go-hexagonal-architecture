package model

import (
	"path/filepath"
	"strings"
	"time"

	domainerrors "github.com/linkai/go-chatbot-api/domain/errors"
)

type Chatbot struct {
	ID                   string
	Name                 string
	Description          string
	SystemPrompt         string
	TotalKnowledgeTokens int
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type KnowledgeFileKind string

const (
	KnowledgeFileKindPDF  KnowledgeFileKind = "pdf"
	KnowledgeFileKindTXT  KnowledgeFileKind = "txt"
	KnowledgeFileKindDOCX KnowledgeFileKind = "docx"
	KnowledgeFileKindMD   KnowledgeFileKind = "md"
)

type KnowledgeFile struct {
	ID              string
	ChatbotID       string
	Name            string
	ContentType     string
	SizeBytes       int64
	Kind            KnowledgeFileKind
	StoragePath     string
	ExtractedText   string
	ExtractedTokens int
	CreatedAt       time.Time
}

type CompanyChatbotGrant struct {
	CompanyID string
	ChatbotID string
	CreatedAt time.Time
}

func NewChatbot(id, name, description, systemPrompt string, now time.Time) (Chatbot, error) {
	bot := Chatbot{
		ID:           strings.TrimSpace(id),
		Name:         strings.TrimSpace(name),
		Description:  strings.TrimSpace(description),
		SystemPrompt: strings.TrimSpace(systemPrompt),
		CreatedAt:    now.UTC(),
		UpdatedAt:    now.UTC(),
	}
	if err := bot.Validate(); err != nil {
		return Chatbot{}, err
	}
	return bot, nil
}

func (c Chatbot) Validate() error {
	if c.ID == "" {
		return &domainerrors.ValidationError{Field: "id", Message: "is required"}
	}
	if c.Name == "" {
		return &domainerrors.ValidationError{Field: "name", Message: "is required"}
	}
	if c.SystemPrompt == "" {
		return &domainerrors.ValidationError{Field: "system_prompt", Message: "is required"}
	}
	if c.TotalKnowledgeTokens < 0 {
		return &domainerrors.ValidationError{Field: "total_knowledge_tokens", Message: "must be non-negative"}
	}
	return nil
}

func (c Chatbot) WithUpdatedFields(name, description, systemPrompt string, now time.Time) (Chatbot, error) {
	c.Name = strings.TrimSpace(name)
	c.Description = strings.TrimSpace(description)
	c.SystemPrompt = strings.TrimSpace(systemPrompt)
	c.UpdatedAt = now.UTC()
	return c, c.Validate()
}

func DetectKnowledgeFileKind(name string) (KnowledgeFileKind, error) {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".pdf":
		return KnowledgeFileKindPDF, nil
	case ".txt":
		return KnowledgeFileKindTXT, nil
	case ".docx":
		return KnowledgeFileKindDOCX, nil
	case ".md":
		return KnowledgeFileKindMD, nil
	default:
		return "", domainerrors.ErrUnsupportedFileType
	}
}

func NewKnowledgeFile(id, chatbotID, name, contentType, storagePath, extractedText string, sizeBytes int64, extractedTokens int, now time.Time) (KnowledgeFile, error) {
	kind, err := DetectKnowledgeFileKind(name)
	if err != nil {
		return KnowledgeFile{}, err
	}
	file := KnowledgeFile{
		ID:              strings.TrimSpace(id),
		ChatbotID:       strings.TrimSpace(chatbotID),
		Name:            strings.TrimSpace(name),
		ContentType:     strings.TrimSpace(contentType),
		SizeBytes:       sizeBytes,
		Kind:            kind,
		StoragePath:     strings.TrimSpace(storagePath),
		ExtractedText:   strings.TrimSpace(extractedText),
		ExtractedTokens: extractedTokens,
		CreatedAt:       now.UTC(),
	}
	if err := file.Validate(); err != nil {
		return KnowledgeFile{}, err
	}
	return file, nil
}

func (f KnowledgeFile) Validate() error {
	if f.ID == "" {
		return &domainerrors.ValidationError{Field: "id", Message: "is required"}
	}
	if f.ChatbotID == "" {
		return &domainerrors.ValidationError{Field: "chatbot_id", Message: "is required"}
	}
	if f.Name == "" {
		return &domainerrors.ValidationError{Field: "name", Message: "is required"}
	}
	if f.StoragePath == "" {
		return &domainerrors.ValidationError{Field: "storage_path", Message: "is required"}
	}
	if f.SizeBytes <= 0 {
		return &domainerrors.ValidationError{Field: "size_bytes", Message: "must be positive"}
	}
	if f.ExtractedText == "" {
		return &domainerrors.ValidationError{Field: "extracted_text", Message: "must not be empty"}
	}
	if f.ExtractedTokens <= 0 {
		return &domainerrors.ValidationError{Field: "extracted_tokens", Message: "must be positive"}
	}
	return nil
}

func ValidateKnowledgeTokenBudget(totalTokens, maxTokens int) error {
	if totalTokens > maxTokens {
		return domainerrors.ErrKnowledgeLimitExceeded
	}
	return nil
}
