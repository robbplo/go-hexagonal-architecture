package postgres

import (
	"context"
	"fmt"

	"github.com/linkai/go-chatbot-api/domain/model"
	"github.com/linkai/go-chatbot-api/domain/ports"
)

type ChatbotRepo struct {
	q *Queries
}

var _ ports.ChatbotRepository = (*ChatbotRepo)(nil)

func NewChatbotRepo(db DBTX) *ChatbotRepo {
	return &ChatbotRepo{q: New(db)}
}

func (r *ChatbotRepo) Create(ctx context.Context, chatbot model.Chatbot) error {
	if err := r.q.InsertChatbot(ctx, toChatbotRow(chatbot)); err != nil {
		return fmt.Errorf("create chatbot %s: %w", chatbot.ID, mapError(err))
	}
	return nil
}

func (r *ChatbotRepo) GetByID(ctx context.Context, id string) (model.Chatbot, error) {
	row, err := r.q.GetChatbotByID(ctx, id)
	if err != nil {
		return model.Chatbot{}, fmt.Errorf("get chatbot %s: %w", id, mapError(err))
	}
	return toDomainChatbot(row), nil
}

func (r *ChatbotRepo) List(ctx context.Context) ([]model.Chatbot, error) {
	rows, err := r.q.ListChatbots(ctx)
	if err != nil {
		return nil, fmt.Errorf("list chatbots: %w", mapError(err))
	}
	out := make([]model.Chatbot, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainChatbot(row))
	}
	return out, nil
}

func (r *ChatbotRepo) ListByCompany(ctx context.Context, companyID string) ([]model.Chatbot, error) {
	rows, err := r.q.ListChatbotsByCompany(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("list chatbots by company: %w", mapError(err))
	}
	out := make([]model.Chatbot, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainChatbot(row))
	}
	return out, nil
}

func (r *ChatbotRepo) Update(ctx context.Context, chatbot model.Chatbot) error {
	if err := r.q.UpdateChatbot(ctx, toChatbotRow(chatbot)); err != nil {
		return fmt.Errorf("update chatbot %s: %w", chatbot.ID, mapError(err))
	}
	return nil
}

func (r *ChatbotRepo) Delete(ctx context.Context, id string) error {
	if err := r.q.DeleteChatbot(ctx, id); err != nil {
		return fmt.Errorf("delete chatbot %s: %w", id, mapError(err))
	}
	return nil
}

func (r *ChatbotRepo) ListFiles(ctx context.Context, chatbotID string) ([]model.KnowledgeFile, error) {
	rows, err := r.q.ListChatbotFiles(ctx, chatbotID)
	if err != nil {
		return nil, fmt.Errorf("list chatbot files %s: %w", chatbotID, mapError(err))
	}
	out := make([]model.KnowledgeFile, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainKnowledgeFile(row))
	}
	return out, nil
}

func (r *ChatbotRepo) CountFiles(ctx context.Context, chatbotID string) (int, error) {
	count, err := r.q.CountChatbotFiles(ctx, chatbotID)
	if err != nil {
		return 0, fmt.Errorf("count chatbot files %s: %w", chatbotID, mapError(err))
	}
	return count, nil
}

func (r *ChatbotRepo) AddFile(ctx context.Context, file model.KnowledgeFile) error {
	if err := r.q.InsertChatbotFile(ctx, toKnowledgeFileRow(file)); err != nil {
		return fmt.Errorf("add chatbot file %s: %w", file.ID, mapError(err))
	}
	return nil
}

func (r *ChatbotRepo) GetFile(ctx context.Context, chatbotID, fileID string) (model.KnowledgeFile, error) {
	row, err := r.q.GetChatbotFile(ctx, chatbotID, fileID)
	if err != nil {
		return model.KnowledgeFile{}, fmt.Errorf("get chatbot file %s: %w", fileID, mapError(err))
	}
	return toDomainKnowledgeFile(row), nil
}

func (r *ChatbotRepo) DeleteFile(ctx context.Context, chatbotID, fileID string) (model.KnowledgeFile, error) {
	row, err := r.q.DeleteChatbotFile(ctx, chatbotID, fileID)
	if err != nil {
		return model.KnowledgeFile{}, fmt.Errorf("delete chatbot file %s: %w", fileID, mapError(err))
	}
	return toDomainKnowledgeFile(row), nil
}

func (r *ChatbotRepo) SetTotalKnowledgeTokens(ctx context.Context, chatbotID string, totalTokens int) error {
	if err := r.q.SetChatbotKnowledgeTokens(ctx, chatbotID, totalTokens); err != nil {
		return fmt.Errorf("set chatbot tokens %s: %w", chatbotID, mapError(err))
	}
	return nil
}
