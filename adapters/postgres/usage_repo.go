package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainerrors "github.com/linkai/go-chatbot-api/domain/errors"
	"github.com/linkai/go-chatbot-api/domain/model"
	"github.com/linkai/go-chatbot-api/domain/ports"
)

type UsageRepo struct {
	q *Queries
}

var _ ports.CompanyUsageRepository = (*UsageRepo)(nil)

func NewUsageRepo(db DBTX) *UsageRepo {
	return &UsageRepo{q: New(db)}
}

func (r *UsageRepo) GetOrCreate(ctx context.Context, companyID string, monthStart time.Time, budgetTokens int64) (model.CompanyMonthUsage, error) {
	row, err := r.q.GetCompanyMonthUsage(ctx, companyID, monthStart)
	if err == nil {
		return toDomainUsage(row), nil
	}
	mapped := mapError(err)
	if !errors.Is(mapped, domainerrors.ErrNotFound) {
		return model.CompanyMonthUsage{}, fmt.Errorf("get usage %s: %w", companyID, mapped)
	}
	usage := model.CompanyMonthUsage{
		CompanyID:    companyID,
		MonthStart:   monthStart,
		BudgetTokens: budgetTokens,
		UpdatedAt:    time.Now().UTC(),
	}
	if err := r.q.UpsertCompanyMonthUsage(ctx, toUsageRow(usage)); err != nil {
		return model.CompanyMonthUsage{}, fmt.Errorf("create usage %s: %w", companyID, mapError(err))
	}
	return usage, nil
}

func (r *UsageRepo) AddEvent(ctx context.Context, event model.TokenUsageEvent, budgetTokens int64) (model.CompanyMonthUsage, error) {
	tx, err := beginTx(ctx, r.q.db)
	if err != nil {
		return model.CompanyMonthUsage{}, fmt.Errorf("add usage event: %w", err)
	}
	defer tx.Rollback(ctx)

	q := New(tx)
	current, err := q.GetCompanyMonthUsage(ctx, event.CompanyID, event.MonthStart)
	if err != nil {
		mapped := mapError(err)
		if !errors.Is(mapped, domainerrors.ErrNotFound) {
			return model.CompanyMonthUsage{}, fmt.Errorf("add usage event: %w", mapped)
		}
		current = CompanyMonthUsage{
			CompanyID:    event.CompanyID,
			MonthStart:   event.MonthStart,
			BudgetTokens: budgetTokens,
			UpdatedAt:    event.CreatedAt,
		}
	}

	next := toDomainUsage(current)
	next.BudgetTokens = budgetTokens
	next.InputTokens += int64(event.InputTokens)
	next.OutputTokens += int64(event.OutputTokens)
	next.UpdatedAt = event.CreatedAt.UTC()

	if err := q.UpsertCompanyMonthUsage(ctx, toUsageRow(next)); err != nil {
		return model.CompanyMonthUsage{}, fmt.Errorf("add usage event: %w", mapError(err))
	}
	if err := q.InsertTokenUsageEvent(ctx, event.CompanyID, event.ConversationID, event.UserID, event.ChatbotID, event.AssistantMsgID, event.MonthStart, event.InputTokens, event.OutputTokens, event.ID, event.CreatedAt); err != nil {
		return model.CompanyMonthUsage{}, fmt.Errorf("add usage event: %w", mapError(err))
	}
	if err := tx.Commit(ctx); err != nil {
		return model.CompanyMonthUsage{}, fmt.Errorf("add usage event: %w", err)
	}
	return next, nil
}

func (r *UsageRepo) AdjustCurrentMonth(ctx context.Context, companyID string, monthStart time.Time, budgetTokens int64, delta int64) (model.CompanyMonthUsage, error) {
	current, err := r.GetOrCreate(ctx, companyID, monthStart, budgetTokens)
	if err != nil {
		return model.CompanyMonthUsage{}, err
	}
	current, err = current.WithAdjustment(delta, time.Now())
	if err != nil {
		return model.CompanyMonthUsage{}, fmt.Errorf("adjust usage %s: %w", companyID, err)
	}
	current.BudgetTokens = budgetTokens
	if err := r.q.UpsertCompanyMonthUsage(ctx, toUsageRow(current)); err != nil {
		return model.CompanyMonthUsage{}, fmt.Errorf("adjust usage %s: %w", companyID, mapError(err))
	}
	return current, nil
}

func (r *UsageRepo) ResetCurrentMonth(ctx context.Context, companyID string, monthStart time.Time, budgetTokens int64) (model.CompanyMonthUsage, error) {
	current, err := r.GetOrCreate(ctx, companyID, monthStart, budgetTokens)
	if err != nil {
		return model.CompanyMonthUsage{}, err
	}
	current = current.Reset(time.Now())
	current.BudgetTokens = budgetTokens
	if err := r.q.UpsertCompanyMonthUsage(ctx, toUsageRow(current)); err != nil {
		return model.CompanyMonthUsage{}, fmt.Errorf("reset usage %s: %w", companyID, mapError(err))
	}
	return current, nil
}
