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

type CompanyService struct {
	companies  ports.CompanyRepository
	users      ports.UserRepository
	usage      ports.CompanyUsageRepository
	identities ports.IdentityAdmin
	clock      ports.Clock
	ids        ports.IDGenerator
	logger     ports.Logger
}

type CreateCompanyCommand struct {
	Name               string
	MonthlyTokenBudget int64
}

type UpdateCompanyCommand struct {
	CompanyID          string
	Name               string
	MonthlyTokenBudget int64
}

func NewCompanyService(
	companies ports.CompanyRepository,
	users ports.UserRepository,
	usage ports.CompanyUsageRepository,
	identities ports.IdentityAdmin,
	clock ports.Clock,
	ids ports.IDGenerator,
	logger ports.Logger,
) *CompanyService {
	if companies == nil {
		panic("CompanyService: companies is required")
	}
	if users == nil {
		panic("CompanyService: users is required")
	}
	if usage == nil {
		panic("CompanyService: usage is required")
	}
	if identities == nil {
		panic("CompanyService: identities is required")
	}
	if clock == nil {
		panic("CompanyService: clock is required")
	}
	if ids == nil {
		panic("CompanyService: ids is required")
	}
	if logger == nil {
		panic("CompanyService: logger is required")
	}
	return &CompanyService{
		companies:  companies,
		users:      users,
		usage:      usage,
		identities: identities,
		clock:      clock,
		ids:        ids,
		logger:     logger,
	}
}

func (s *CompanyService) Create(ctx context.Context, cmd CreateCompanyCommand) (model.Company, error) {
	company, err := model.NewCompany(s.ids.NewID("cmp"), strings.TrimSpace(cmd.Name), cmd.MonthlyTokenBudget, s.clock.Now())
	if err != nil {
		return model.Company{}, fmt.Errorf("create company: %w", err)
	}
	if err := s.companies.Create(ctx, company); err != nil {
		return model.Company{}, fmt.Errorf("create company: %w", err)
	}
	s.logger.Info(ctx, "company created", "company_id", company.ID)
	return company, nil
}

func (s *CompanyService) List(ctx context.Context) ([]model.Company, error) {
	companies, err := s.companies.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list companies: %w", err)
	}
	return companies, nil
}

func (s *CompanyService) Update(ctx context.Context, cmd UpdateCompanyCommand) (model.Company, error) {
	company, err := s.companies.GetByID(ctx, strings.TrimSpace(cmd.CompanyID))
	if err != nil {
		return model.Company{}, fmt.Errorf("update company: %w", err)
	}
	company, err = company.WithUpdatedFields(cmd.Name, cmd.MonthlyTokenBudget, s.clock.Now())
	if err != nil {
		return model.Company{}, fmt.Errorf("update company: %w", err)
	}
	if err := s.companies.Update(ctx, company); err != nil {
		return model.Company{}, fmt.Errorf("update company: %w", err)
	}
	s.logger.Info(ctx, "company updated", "company_id", company.ID)
	return company, nil
}

func (s *CompanyService) Delete(ctx context.Context, companyID string) error {
	companyID = strings.TrimSpace(companyID)
	activeCount, err := s.users.CountActiveByCompany(ctx, companyID)
	if err != nil {
		return fmt.Errorf("delete company: %w", err)
	}
	if err := model.ValidateCompanyDeletion(activeCount); err != nil {
		return fmt.Errorf("delete company: %w", err)
	}

	remainingUsers, err := s.users.ListByCompanyAndStatuses(ctx, companyID, model.UserStatusInvited, model.UserStatusDisabled)
	if err != nil {
		return fmt.Errorf("delete company: %w", err)
	}
	for _, user := range remainingUsers {
		if err := s.identities.DeleteUser(ctx, user.ID); err != nil && !errors.Is(err, domainerrors.ErrNotFound) {
			return fmt.Errorf("delete company auth user %s: %w", user.ID, err)
		}
		if err := s.users.Delete(ctx, user.ID); err != nil && !errors.Is(err, domainerrors.ErrNotFound) {
			return fmt.Errorf("delete company profile %s: %w", user.ID, err)
		}
	}

	if err := s.companies.Delete(ctx, companyID); err != nil {
		return fmt.Errorf("delete company: %w", err)
	}
	s.logger.Info(ctx, "company deleted", "company_id", companyID)
	return nil
}

func (s *CompanyService) GetCurrentUsage(ctx context.Context, companyID string) (model.CompanyMonthUsage, error) {
	company, err := s.companies.GetByID(ctx, strings.TrimSpace(companyID))
	if err != nil {
		return model.CompanyMonthUsage{}, fmt.Errorf("get company usage: %w", err)
	}
	usage, err := s.usage.GetOrCreate(ctx, company.ID, model.NormalizeMonth(s.clock.Now()), company.MonthlyTokenBudget)
	if err != nil {
		return model.CompanyMonthUsage{}, fmt.Errorf("get company usage: %w", err)
	}
	return usage, nil
}

func (s *CompanyService) AdjustCurrentUsage(ctx context.Context, companyID string, delta int64) (model.CompanyMonthUsage, error) {
	company, err := s.companies.GetByID(ctx, strings.TrimSpace(companyID))
	if err != nil {
		return model.CompanyMonthUsage{}, fmt.Errorf("adjust company usage: %w", err)
	}
	usage, err := s.usage.AdjustCurrentMonth(ctx, company.ID, model.NormalizeMonth(s.clock.Now()), company.MonthlyTokenBudget, delta)
	if err != nil {
		return model.CompanyMonthUsage{}, fmt.Errorf("adjust company usage: %w", err)
	}
	s.logger.Info(ctx, "company usage adjusted", "company_id", company.ID, "delta", delta)
	return usage, nil
}

func (s *CompanyService) ResetCurrentUsage(ctx context.Context, companyID string) (model.CompanyMonthUsage, error) {
	company, err := s.companies.GetByID(ctx, strings.TrimSpace(companyID))
	if err != nil {
		return model.CompanyMonthUsage{}, fmt.Errorf("reset company usage: %w", err)
	}
	usage, err := s.usage.ResetCurrentMonth(ctx, company.ID, model.NormalizeMonth(s.clock.Now()), company.MonthlyTokenBudget)
	if err != nil {
		return model.CompanyMonthUsage{}, fmt.Errorf("reset company usage: %w", err)
	}
	s.logger.Info(ctx, "company usage reset", "company_id", company.ID)
	return usage, nil
}
