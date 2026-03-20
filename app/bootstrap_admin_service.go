package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domainerrors "github.com/linkai/go-chatbot-api/domain/errors"
	"github.com/linkai/go-chatbot-api/domain/model"
	"github.com/linkai/go-chatbot-api/domain/ports"
)

type BootstrapAdminService struct {
	users      ports.UserRepository
	identities ports.IdentityAdmin
	clock      ports.Clock
	logger     ports.Logger
}

func NewBootstrapAdminService(users ports.UserRepository, identities ports.IdentityAdmin, clock ports.Clock, logger ports.Logger) *BootstrapAdminService {
	if users == nil {
		panic("BootstrapAdminService: users is required")
	}
	if identities == nil {
		panic("BootstrapAdminService: identities is required")
	}
	if clock == nil {
		panic("BootstrapAdminService: clock is required")
	}
	if logger == nil {
		panic("BootstrapAdminService: logger is required")
	}
	return &BootstrapAdminService{
		users:      users,
		identities: identities,
		clock:      clock,
		logger:     logger,
	}
}

func (s *BootstrapAdminService) EnsureAdmin(ctx context.Context, email, password string) (model.UserProfile, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	profile, err := s.users.GetByEmail(ctx, email)
	if err == nil {
		if profile.Role != model.RoleAdmin {
			return model.UserProfile{}, fmt.Errorf("ensure admin: %w", domainerrors.ErrConflict)
		}
		return profile, nil
	}
	if !errors.Is(err, domainerrors.ErrNotFound) {
		return model.UserProfile{}, fmt.Errorf("ensure admin: %w", err)
	}

	identity, err := s.identities.CreateUser(ctx, email, password, true)
	if err != nil {
		return model.UserProfile{}, fmt.Errorf("ensure admin: %w", err)
	}
	profile, err = model.NewAdminProfile(identity.ID, identity.Email, s.clock.Now())
	if err != nil {
		return model.UserProfile{}, fmt.Errorf("ensure admin: %w", err)
	}
	if err := s.users.Create(ctx, profile); err != nil {
		return model.UserProfile{}, fmt.Errorf("ensure admin: %w", err)
	}
	s.logger.Info(ctx, "admin bootstrapped", "user_id", profile.ID)
	return profile, nil
}
