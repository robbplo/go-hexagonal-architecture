package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/linkai/go-chatbot-api/domain/model"
	"github.com/linkai/go-chatbot-api/domain/ports"
)

type UserRepo struct {
	db DBTX
	q  *Queries
}

var _ ports.UserRepository = (*UserRepo)(nil)

func NewUserRepo(db DBTX) *UserRepo {
	return &UserRepo{
		db: db,
		q:  New(db),
	}
}

func (r *UserRepo) Create(ctx context.Context, profile model.UserProfile) error {
	if err := r.q.InsertUserProfile(ctx, toUserProfileRow(profile)); err != nil {
		return fmt.Errorf("create user %s: %w", profile.ID, mapError(err))
	}
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (model.UserProfile, error) {
	row, err := r.q.GetUserProfileByID(ctx, id)
	if err != nil {
		return model.UserProfile{}, fmt.Errorf("get user %s: %w", id, mapError(err))
	}
	return toDomainUserProfile(row), nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (model.UserProfile, error) {
	row, err := r.q.GetUserProfileByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return model.UserProfile{}, fmt.Errorf("get user by email %s: %w", email, mapError(err))
	}
	return toDomainUserProfile(row), nil
}

func (r *UserRepo) List(ctx context.Context, filter model.UserFilter) ([]model.UserProfile, error) {
	query := `
SELECT id, email, role, status, company_id, invited_at, activated_at, disabled_at, created_at, updated_at
FROM user_profiles
WHERE 1 = 1
`
	args := make([]any, 0, 4)
	argPos := 1
	if filter.CompanyID != nil && *filter.CompanyID != "" {
		query += fmt.Sprintf(" AND company_id = $%d", argPos)
		args = append(args, *filter.CompanyID)
		argPos++
	}
	if len(filter.Statuses) > 0 {
		values := make([]string, 0, len(filter.Statuses))
		for _, status := range filter.Statuses {
			values = append(values, fmt.Sprintf("$%d", argPos))
			args = append(args, string(status))
			argPos++
		}
		query += " AND status IN (" + strings.Join(values, ", ") + ")"
	}
	query += " ORDER BY email ASC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", mapError(err))
	}
	defer rows.Close()

	var profiles []model.UserProfile
	for rows.Next() {
		var row UserProfile
		if err := rows.Scan(&row.ID, &row.Email, &row.Role, &row.Status, &row.CompanyID, &row.InvitedAt, &row.ActivatedAt, &row.DisabledAt, &row.CreatedAt, &row.UpdatedAt); err != nil {
			return nil, fmt.Errorf("list users: %w", err)
		}
		profiles = append(profiles, toDomainUserProfile(row))
	}
	return profiles, rows.Err()
}

func (r *UserRepo) Update(ctx context.Context, profile model.UserProfile) error {
	if err := r.q.UpdateUserProfile(ctx, toUserProfileRow(profile)); err != nil {
		return fmt.Errorf("update user %s: %w", profile.ID, mapError(err))
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	if err := r.q.DeleteUserProfile(ctx, id); err != nil {
		return fmt.Errorf("delete user %s: %w", id, mapError(err))
	}
	return nil
}

func (r *UserRepo) CountActiveByCompany(ctx context.Context, companyID string) (int, error) {
	row := r.db.QueryRow(ctx, `SELECT count(*) FROM user_profiles WHERE company_id = $1 AND status = 'active'`, companyID)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count active users for company %s: %w", companyID, mapError(err))
	}
	return count, nil
}

func (r *UserRepo) ListByCompanyAndStatuses(ctx context.Context, companyID string, statuses ...model.UserStatus) ([]model.UserProfile, error) {
	if len(statuses) == 0 {
		return nil, nil
	}
	placeholders := make([]string, 0, len(statuses))
	args := make([]any, 0, len(statuses)+1)
	args = append(args, companyID)
	for index, status := range statuses {
		placeholders = append(placeholders, fmt.Sprintf("$%d", index+2))
		args = append(args, string(status))
	}
	query := `
SELECT id, email, role, status, company_id, invited_at, activated_at, disabled_at, created_at, updated_at
FROM user_profiles
WHERE company_id = $1 AND status IN (` + strings.Join(placeholders, ", ") + `)
ORDER BY email ASC
`
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list users by company and status: %w", mapError(err))
	}
	defer rows.Close()

	var out []model.UserProfile
	for rows.Next() {
		var row UserProfile
		if err := rows.Scan(&row.ID, &row.Email, &row.Role, &row.Status, &row.CompanyID, &row.InvitedAt, &row.ActivatedAt, &row.DisabledAt, &row.CreatedAt, &row.UpdatedAt); err != nil {
			return nil, fmt.Errorf("list users by company and status: %w", err)
		}
		out = append(out, toDomainUserProfile(row))
	}
	if err := rows.Err(); err != nil && err != pgx.ErrNoRows {
		return nil, err
	}
	return out, nil
}
