package app

import (
	"context"
	"fmt"

	"github.com/linkai/go-chatbot-api/domain/model"
	"github.com/linkai/go-chatbot-api/domain/ports"
)

type CatalogService struct {
	chatbots ports.ChatbotRepository
}

func NewCatalogService(chatbots ports.ChatbotRepository) *CatalogService {
	if chatbots == nil {
		panic("CatalogService: chatbots is required")
	}
	return &CatalogService{chatbots: chatbots}
}

func (s *CatalogService) ListForUser(ctx context.Context, user model.UserProfile) ([]model.Chatbot, error) {
	if user.CompanyID == nil {
		return nil, nil
	}
	chatbots, err := s.chatbots.ListByCompany(ctx, *user.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("list available chatbots: %w", err)
	}
	return chatbots, nil
}
