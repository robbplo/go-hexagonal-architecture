package app

import (
	"context"
	"testing"
	"time"

	domainerrors "github.com/linkai/go-chatbot-api/domain/errors"
	"github.com/linkai/go-chatbot-api/domain/model"
	"github.com/linkai/go-chatbot-api/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestChatService_SendMessageBudgetExceeded(t *testing.T) {
	companies := mocks.NewMockCompanyRepository(t)
	chatbots := mocks.NewMockChatbotRepository(t)
	grants := mocks.NewMockGrantRepository(t)
	conversations := mocks.NewMockConversationRepository(t)
	usage := mocks.NewMockCompanyUsageRepository(t)
	ai := mocks.NewMockAIChatClient(t)
	tokenCounter := mocks.NewMockTokenCounter(t)
	clock := mocks.NewMockClock(t)
	ids := mocks.NewMockIDGenerator(t)
	logger := mocks.NewMockLogger(t)

	grants.On("CompanyHasAccess", mock.Anything, "cmp_1", "bot_1").Return(true, nil)
	companies.On("GetByID", mock.Anything, "cmp_1").Return(model.Company{
		ID:                 "cmp_1",
		Name:               "Acme",
		MonthlyTokenBudget: 100,
	}, nil)
	clock.On("Now").Return(time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC))
	usage.On("GetOrCreate", mock.Anything, "cmp_1", model.NormalizeMonth(time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC)), int64(100)).Return(model.CompanyMonthUsage{
		CompanyID:              "cmp_1",
		BudgetTokens:           100,
		InputTokens:            60,
		OutputTokens:           40,
		ManualAdjustmentTokens: 0,
	}, nil)

	service := NewChatService(companies, chatbots, grants, conversations, usage, ai, tokenCounter, clock, ids, logger, "gpt-4.1-mini", 1000)
	_, err := service.SendMessage(context.Background(), model.UserProfile{
		ID:        "usr_1",
		Email:     "user@example.com",
		Role:      model.RoleUser,
		Status:    model.UserStatusActive,
		CompanyID: ptr("cmp_1"),
	}, "bot_1", "hello")

	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrTokenBudgetExceeded)
}

func ptr(value string) *string {
	return &value
}
