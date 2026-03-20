package app

import (
	"context"
	"errors"
	"testing"
	"time"

	domainerrors "github.com/linkai/go-chatbot-api/domain/errors"
	"github.com/linkai/go-chatbot-api/domain/model"
	"github.com/linkai/go-chatbot-api/domain/ports"
	"github.com/linkai/go-chatbot-api/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserService_Invite(t *testing.T) {
	users := mocks.NewMockUserRepository(t)
	companies := mocks.NewMockCompanyRepository(t)
	identities := mocks.NewMockIdentityAdmin(t)
	clock := mocks.NewMockClock(t)
	logger := mocks.NewMockLogger(t)
	now := time.Date(2026, 3, 13, 12, 0, 0, 0, time.UTC)

	companies.On("GetByID", mock.Anything, "cmp_1").Return(model.Company{ID: "cmp_1", Name: "Acme"}, nil)
	users.On("GetByEmail", mock.Anything, "user@example.com").Return(model.UserProfile{}, domainerrors.ErrNotFound)
	identities.On("InviteUser", mock.Anything, "user@example.com", "https://app.example.com/invite").Return(ports.IdentityUser{
		ID:    "usr_1",
		Email: "user@example.com",
	}, nil)
	clock.On("Now").Return(now)
	users.On("Create", mock.Anything, mock.MatchedBy(func(profile model.UserProfile) bool {
		return profile.ID == "usr_1" && profile.Status == model.UserStatusInvited
	})).Return(nil)
	logger.On("Info", mock.Anything, "user invited", "user_id", "usr_1", "company_id", "cmp_1").Return()

	service := NewUserService(users, companies, identities, clock, logger)
	profile, err := service.Invite(context.Background(), InviteUserCommand{
		Email:       "user@example.com",
		CompanyID:   "cmp_1",
		RedirectURL: "https://app.example.com/invite",
	})

	require.NoError(t, err)
	require.Equal(t, "usr_1", profile.ID)
	require.False(t, errors.Is(err, domainerrors.ErrNotFound))
}
