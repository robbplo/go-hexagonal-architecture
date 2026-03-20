package model

import (
	"time"

	domainerrors "github.com/linkai/go-chatbot-api/domain/errors"
)

type CompanyMonthUsage struct {
	CompanyID              string
	MonthStart             time.Time
	BudgetTokens           int64
	InputTokens            int64
	OutputTokens           int64
	ManualAdjustmentTokens int64
	UpdatedAt              time.Time
}

type TokenUsageEvent struct {
	ID             string
	CompanyID      string
	ConversationID string
	UserID         string
	ChatbotID      string
	AssistantMsgID string
	MonthStart     time.Time
	InputTokens    int
	OutputTokens   int
	CreatedAt      time.Time
}

func NormalizeMonth(now time.Time) time.Time {
	value := now.UTC()
	return time.Date(value.Year(), value.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func (u CompanyMonthUsage) Validate() error {
	if u.CompanyID == "" {
		return &domainerrors.ValidationError{Field: "company_id", Message: "is required"}
	}
	if u.BudgetTokens < 0 {
		return &domainerrors.ValidationError{Field: "budget_tokens", Message: "must be non-negative"}
	}
	return nil
}

func (u CompanyMonthUsage) EffectiveUsage() int64 {
	return u.InputTokens + u.OutputTokens + u.ManualAdjustmentTokens
}

func (u CompanyMonthUsage) RemainingTokens() int64 {
	remaining := u.BudgetTokens - u.EffectiveUsage()
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (u CompanyMonthUsage) ExceedsBudget() bool {
	return u.EffectiveUsage() >= u.BudgetTokens && u.BudgetTokens > 0
}

func (u CompanyMonthUsage) WithAdjustment(delta int64, now time.Time) (CompanyMonthUsage, error) {
	next := u
	next.ManualAdjustmentTokens += delta
	if next.EffectiveUsage() < 0 {
		return CompanyMonthUsage{}, &domainerrors.ValidationError{Field: "manual_adjustment_tokens", Message: "must not make effective usage negative"}
	}
	next.UpdatedAt = now.UTC()
	return next, nil
}

func (u CompanyMonthUsage) Reset(now time.Time) CompanyMonthUsage {
	next := u
	next.ManualAdjustmentTokens = -(u.InputTokens + u.OutputTokens)
	next.UpdatedAt = now.UTC()
	return next
}
