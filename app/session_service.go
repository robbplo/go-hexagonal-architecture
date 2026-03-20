package app

import (
	"context"
	"fmt"
	"strings"

	domainerrors "github.com/linkai/go-chatbot-api/domain/errors"
	"github.com/linkai/go-chatbot-api/domain/model"
	"github.com/linkai/go-chatbot-api/domain/ports"
)

type SessionService struct {
	users  ports.UserRepository
	clock  ports.Clock
	logger ports.Logger
}

func NewSessionService(users ports.UserRepository, clock ports.Clock, logger ports.Logger) *SessionService {
	if users == nil {
		panic("SessionService: users is required")
	}
	if clock == nil {
		panic("SessionService: clock is required")
	}
	if logger == nil {
		panic("SessionService: logger is required")
	}
	return &SessionService{
		users:  users,
		clock:  clock,
		logger: logger,
	}
}

func (s *SessionService) ResolveAuthenticatedProfile(ctx context.Context, authUser ports.AuthenticatedUser) (model.UserProfile, error) {
	profile, err := s.users.GetByID(ctx, strings.TrimSpace(authUser.ID))
	if err != nil {
		return model.UserProfile{}, fmt.Errorf("resolve session profile: %w", err)
	}
	if profile.Status == model.UserStatusDisabled {
		return model.UserProfile{}, fmt.Errorf("resolve session profile: %w", domainerrors.ErrUnauthorized)
	}
	if profile.Status == model.UserStatusInvited {
		profile, err = profile.Activate(s.clock.Now())
		if err != nil {
			return model.UserProfile{}, fmt.Errorf("resolve session profile: %w", err)
		}
		if err := s.users.Update(ctx, profile); err != nil {
			return model.UserProfile{}, fmt.Errorf("resolve session profile: %w", err)
		}
		s.logger.Info(ctx, "user activated", "user_id", profile.ID)
	}
	return profile, nil
}
