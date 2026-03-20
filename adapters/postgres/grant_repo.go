package postgres

import (
	"context"
	"fmt"

	"github.com/linkai/go-chatbot-api/domain/model"
	"github.com/linkai/go-chatbot-api/domain/ports"
)

type GrantRepo struct {
	q *Queries
}

var _ ports.GrantRepository = (*GrantRepo)(nil)

func NewGrantRepo(db DBTX) *GrantRepo {
	return &GrantRepo{q: New(db)}
}

func (r *GrantRepo) Grant(ctx context.Context, grant model.CompanyChatbotGrant) error {
	if err := r.q.GrantAccess(ctx, grant.CompanyID, grant.ChatbotID, grant.CreatedAt); err != nil {
		return fmt.Errorf("grant access company=%s chatbot=%s: %w", grant.CompanyID, grant.ChatbotID, mapError(err))
	}
	return nil
}

func (r *GrantRepo) Revoke(ctx context.Context, companyID, chatbotID string) error {
	if err := r.q.RevokeAccess(ctx, companyID, chatbotID); err != nil {
		return fmt.Errorf("revoke access company=%s chatbot=%s: %w", companyID, chatbotID, mapError(err))
	}
	return nil
}

func (r *GrantRepo) CompanyHasAccess(ctx context.Context, companyID, chatbotID string) (bool, error) {
	value, err := r.q.CompanyHasAccess(ctx, companyID, chatbotID)
	if err != nil {
		return false, fmt.Errorf("check access company=%s chatbot=%s: %w", companyID, chatbotID, mapError(err))
	}
	return value, nil
}

func (r *GrantRepo) ListCompaniesByChatbot(ctx context.Context, chatbotID string) ([]model.Company, error) {
	rows, err := r.q.ListCompaniesByChatbot(ctx, chatbotID)
	if err != nil {
		return nil, fmt.Errorf("list companies by chatbot %s: %w", chatbotID, mapError(err))
	}
	out := make([]model.Company, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainCompany(row))
	}
	return out, nil
}
