package model

import (
	"fmt"
	"strings"
	"time"

	domainerrors "github.com/linkai/go-chatbot-api/domain/errors"
)

type Company struct {
	ID                 string
	Name               string
	MonthlyTokenBudget int64
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func NewCompany(id, name string, monthlyTokenBudget int64, now time.Time) (Company, error) {
	company := Company{
		ID:                 strings.TrimSpace(id),
		Name:               strings.TrimSpace(name),
		MonthlyTokenBudget: monthlyTokenBudget,
		CreatedAt:          now.UTC(),
		UpdatedAt:          now.UTC(),
	}
	if err := company.Validate(); err != nil {
		return Company{}, err
	}
	return company, nil
}

func (c Company) Validate() error {
	if c.ID == "" {
		return &domainerrors.ValidationError{Field: "id", Message: "is required"}
	}
	if c.Name == "" {
		return &domainerrors.ValidationError{Field: "name", Message: "is required"}
	}
	if c.MonthlyTokenBudget < 0 {
		return &domainerrors.ValidationError{Field: "monthly_token_budget", Message: "must be non-negative"}
	}
	return nil
}

func (c Company) WithUpdatedFields(name string, monthlyTokenBudget int64, now time.Time) (Company, error) {
	c.Name = strings.TrimSpace(name)
	c.MonthlyTokenBudget = monthlyTokenBudget
	c.UpdatedAt = now.UTC()
	if err := c.Validate(); err != nil {
		return Company{}, err
	}
	return c, nil
}

func ValidateCompanyDeletion(activeUsers int) error {
	if activeUsers > 0 {
		return fmt.Errorf("company has active users: %w", domainerrors.ErrConflict)
	}
	return nil
}
