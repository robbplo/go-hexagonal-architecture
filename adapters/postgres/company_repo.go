package postgres

import (
	"context"
	"fmt"

	"github.com/linkai/go-chatbot-api/domain/model"
	"github.com/linkai/go-chatbot-api/domain/ports"
)

type CompanyRepo struct {
	q *Queries
}

var _ ports.CompanyRepository = (*CompanyRepo)(nil)

func NewCompanyRepo(db DBTX) *CompanyRepo {
	return &CompanyRepo{q: New(db)}
}

func (r *CompanyRepo) Create(ctx context.Context, company model.Company) error {
	if err := r.q.CreateCompany(ctx, toCompanyRow(company)); err != nil {
		return fmt.Errorf("create company: %w", mapError(err))
	}
	return nil
}

func (r *CompanyRepo) GetByID(ctx context.Context, id string) (model.Company, error) {
	row, err := r.q.GetCompanyByID(ctx, id)
	if err != nil {
		return model.Company{}, fmt.Errorf("get company %s: %w", id, mapError(err))
	}
	return toDomainCompany(row), nil
}

func (r *CompanyRepo) List(ctx context.Context) ([]model.Company, error) {
	rows, err := r.q.ListCompanies(ctx)
	if err != nil {
		return nil, fmt.Errorf("list companies: %w", mapError(err))
	}
	out := make([]model.Company, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainCompany(row))
	}
	return out, nil
}

func (r *CompanyRepo) Update(ctx context.Context, company model.Company) error {
	if err := r.q.UpdateCompany(ctx, toCompanyRow(company)); err != nil {
		return fmt.Errorf("update company %s: %w", company.ID, mapError(err))
	}
	return nil
}

func (r *CompanyRepo) Delete(ctx context.Context, id string) error {
	if err := r.q.DeleteCompany(ctx, id); err != nil {
		return fmt.Errorf("delete company %s: %w", id, mapError(err))
	}
	return nil
}
