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

type UserService struct {
	users      ports.UserRepository
	companies  ports.CompanyRepository
	identities ports.IdentityAdmin
	clock      ports.Clock
	logger     ports.Logger
}

type InviteUserCommand struct {
	Email       string
	CompanyID   string
	RedirectURL string
}

func NewUserService(
	users ports.UserRepository,
	companies ports.CompanyRepository,
	identities ports.IdentityAdmin,
	clock ports.Clock,
	logger ports.Logger,
) *UserService {
	if users == nil {
		panic("UserService: users is required")
	}
	if companies == nil {
		panic("UserService: companies is required")
	}
	if identities == nil {
		panic("UserService: identities is required")
	}
	if clock == nil {
		panic("UserService: clock is required")
	}
	if logger == nil {
		panic("UserService: logger is required")
	}
	return &UserService{
		users:      users,
		companies:  companies,
		identities: identities,
		clock:      clock,
		logger:     logger,
	}
}

func (s *UserService) Invite(ctx context.Context, cmd InviteUserCommand) (model.UserProfile, error) {
	email := strings.ToLower(strings.TrimSpace(cmd.Email))
	companyID := strings.TrimSpace(cmd.CompanyID)

	if _, err := s.companies.GetByID(ctx, companyID); err != nil {
		return model.UserProfile{}, fmt.Errorf("invite user: %w", err)
	}
	if _, err := s.users.GetByEmail(ctx, email); err == nil {
		return model.UserProfile{}, fmt.Errorf("invite user: %w", domainerrors.ErrAlreadyExists)
	} else if !errors.Is(err, domainerrors.ErrNotFound) {
		return model.UserProfile{}, fmt.Errorf("invite user: %w", err)
	}

	identity, err := s.identities.InviteUser(ctx, email, cmd.RedirectURL)
	if err != nil {
		return model.UserProfile{}, fmt.Errorf("invite user: %w", err)
	}
	profile, err := model.NewInvitedUserProfile(identity.ID, identity.Email, companyID, s.clock.Now())
	if err != nil {
		return model.UserProfile{}, fmt.Errorf("invite user: %w", err)
	}
	if err := s.users.Create(ctx, profile); err != nil {
		_ = s.identities.DeleteUser(ctx, identity.ID)
		return model.UserProfile{}, fmt.Errorf("invite user: %w", err)
	}
	s.logger.Info(ctx, "user invited", "user_id", profile.ID, "company_id", companyID)
	return profile, nil
}

func (s *UserService) List(ctx context.Context, filter model.UserFilter) ([]model.UserProfile, error) {
	users, err := s.users.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (s *UserService) Disable(ctx context.Context, userID string) (model.UserProfile, error) {
	profile, err := s.users.GetByID(ctx, strings.TrimSpace(userID))
	if err != nil {
		return model.UserProfile{}, fmt.Errorf("disable user: %w", err)
	}
	if profile.Role == model.RoleAdmin {
		return model.UserProfile{}, fmt.Errorf("disable user: %w", domainerrors.ErrForbidden)
	}
	profile, err = profile.Disable(s.clock.Now())
	if err != nil {
		return model.UserProfile{}, fmt.Errorf("disable user: %w", err)
	}
	if err := s.identities.DisableUser(ctx, profile.ID); err != nil {
		return model.UserProfile{}, fmt.Errorf("disable user: %w", err)
	}
	if err := s.users.Update(ctx, profile); err != nil {
		return model.UserProfile{}, fmt.Errorf("disable user: %w", err)
	}
	s.logger.Info(ctx, "user disabled", "user_id", profile.ID)
	return profile, nil
}

func (s *UserService) Delete(ctx context.Context, userID string) error {
	profile, err := s.users.GetByID(ctx, strings.TrimSpace(userID))
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if profile.Role == model.RoleAdmin {
		return fmt.Errorf("delete user: %w", domainerrors.ErrForbidden)
	}
	if err := s.identities.DeleteUser(ctx, profile.ID); err != nil && !errors.Is(err, domainerrors.ErrNotFound) {
		return fmt.Errorf("delete user: %w", err)
	}
	if err := s.users.Delete(ctx, profile.ID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	s.logger.Info(ctx, "user deleted", "user_id", profile.ID)
	return nil
}
