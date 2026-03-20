package app

import (
	"context"
	"testing"
	"time"

	"github.com/linkai/go-chatbot-api/domain/model"
	"github.com/linkai/go-chatbot-api/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCompanyService_DeleteWithActiveUsers(t *testing.T) {
	companies := mocks.NewMockCompanyRepository(t)
	users := mocks.NewMockUserRepository(t)
	usage := mocks.NewMockCompanyUsageRepository(t)
	identities := mocks.NewMockIdentityAdmin(t)
	clock := mocks.NewMockClock(t)
	ids := mocks.NewMockIDGenerator(t)
	logger := mocks.NewMockLogger(t)

	users.On("CountActiveByCompany", mock.Anything, "cmp_1").Return(1, nil)

	service := NewCompanyService(companies, users, usage, identities, clock, ids, logger)
	err := service.Delete(context.Background(), "cmp_1")

	require.Error(t, err)
	assert.ErrorContains(t, err, "conflict")
}

func TestCompanyService_ResetCurrentUsage(t *testing.T) {
	companies := mocks.NewMockCompanyRepository(t)
	users := mocks.NewMockUserRepository(t)
	usage := mocks.NewMockCompanyUsageRepository(t)
	identities := mocks.NewMockIdentityAdmin(t)
	clock := mocks.NewMockClock(t)
	ids := mocks.NewMockIDGenerator(t)
	logger := mocks.NewMockLogger(t)
	now := time.Date(2026, 3, 13, 10, 0, 0, 0, time.UTC)

	companies.On("GetByID", mock.Anything, "cmp_1").Return(model.Company{
		ID:                 "cmp_1",
		Name:               "Acme",
		MonthlyTokenBudget: 1000,
	}, nil)
	clock.On("Now").Return(now)
	usage.On("ResetCurrentMonth", mock.Anything, "cmp_1", model.NormalizeMonth(now), int64(1000)).Return(model.CompanyMonthUsage{
		CompanyID:    "cmp_1",
		BudgetTokens: 1000,
	}, nil)
	logger.On("Info", mock.Anything, "company usage reset", "company_id", "cmp_1").Return()

	service := NewCompanyService(companies, users, usage, identities, clock, ids, logger)
	got, err := service.ResetCurrentUsage(context.Background(), "cmp_1")

	require.NoError(t, err)
	assert.Equal(t, "cmp_1", got.CompanyID)
}
