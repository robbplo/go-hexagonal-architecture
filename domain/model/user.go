package model

import (
	"net/mail"
	"strings"
	"time"

	domainerrors "github.com/linkai/go-chatbot-api/domain/errors"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type UserStatus string

const (
	UserStatusInvited  UserStatus = "invited"
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
)

type UserProfile struct {
	ID          string
	Email       string
	Role        Role
	Status      UserStatus
	CompanyID   *string
	InvitedAt   *time.Time
	ActivatedAt *time.Time
	DisabledAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UserFilter struct {
	CompanyID *string
	Statuses  []UserStatus
}

func NewInvitedUserProfile(id, email, companyID string, now time.Time) (UserProfile, error) {
	companyID = strings.TrimSpace(companyID)
	invitedAt := now.UTC()
	profile := UserProfile{
		ID:        strings.TrimSpace(id),
		Email:     strings.ToLower(strings.TrimSpace(email)),
		Role:      RoleUser,
		Status:    UserStatusInvited,
		CompanyID: &companyID,
		InvitedAt: &invitedAt,
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC(),
	}
	if err := profile.Validate(); err != nil {
		return UserProfile{}, err
	}
	return profile, nil
}

func NewAdminProfile(id, email string, now time.Time) (UserProfile, error) {
	profile := UserProfile{
		ID:        strings.TrimSpace(id),
		Email:     strings.ToLower(strings.TrimSpace(email)),
		Role:      RoleAdmin,
		Status:    UserStatusActive,
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC(),
	}
	activatedAt := now.UTC()
	profile.ActivatedAt = &activatedAt
	if err := profile.Validate(); err != nil {
		return UserProfile{}, err
	}
	return profile, nil
}

func (u UserProfile) Validate() error {
	if u.ID == "" {
		return &domainerrors.ValidationError{Field: "id", Message: "is required"}
	}
	if u.Email == "" {
		return &domainerrors.ValidationError{Field: "email", Message: "is required"}
	}
	if _, err := mail.ParseAddress(u.Email); err != nil {
		return &domainerrors.ValidationError{Field: "email", Message: "must be a valid email"}
	}
	switch u.Role {
	case RoleAdmin:
		if u.CompanyID != nil {
			return &domainerrors.ValidationError{Field: "company_id", Message: "must be empty for admin users"}
		}
	case RoleUser:
		if u.CompanyID == nil || strings.TrimSpace(*u.CompanyID) == "" {
			return &domainerrors.ValidationError{Field: "company_id", Message: "is required for end users"}
		}
	default:
		return &domainerrors.ValidationError{Field: "role", Message: "is invalid"}
	}
	switch u.Status {
	case UserStatusInvited, UserStatusActive, UserStatusDisabled:
	default:
		return &domainerrors.ValidationError{Field: "status", Message: "is invalid"}
	}
	return nil
}

func (u UserProfile) Activate(now time.Time) (UserProfile, error) {
	if u.Status == UserStatusDisabled {
		return UserProfile{}, domainerrors.ErrUnauthorized
	}
	activatedAt := now.UTC()
	u.Status = UserStatusActive
	u.ActivatedAt = &activatedAt
	u.UpdatedAt = now.UTC()
	return u, u.Validate()
}

func (u UserProfile) Disable(now time.Time) (UserProfile, error) {
	disabledAt := now.UTC()
	u.Status = UserStatusDisabled
	u.DisabledAt = &disabledAt
	u.UpdatedAt = now.UTC()
	return u, u.Validate()
}
