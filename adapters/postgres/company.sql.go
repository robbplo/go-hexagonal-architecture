package postgres

import "context"

const createCompanySQL = `
INSERT INTO companies (id, name, monthly_token_budget, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
`

const getCompanyByIDSQL = `
SELECT id, name, monthly_token_budget, created_at, updated_at
FROM companies
WHERE id = $1
`

const listCompaniesSQL = `
SELECT id, name, monthly_token_budget, created_at, updated_at
FROM companies
ORDER BY name ASC
`

const updateCompanySQL = `
UPDATE companies
SET name = $2, monthly_token_budget = $3, updated_at = $4
WHERE id = $1
`

const deleteCompanySQL = `
DELETE FROM companies
WHERE id = $1
`

func (q *Queries) CreateCompany(ctx context.Context, arg Company) error {
	_, err := q.db.Exec(ctx, createCompanySQL, arg.ID, arg.Name, arg.MonthlyTokenBudget, arg.CreatedAt, arg.UpdatedAt)
	return err
}

func (q *Queries) GetCompanyByID(ctx context.Context, id string) (Company, error) {
	row := q.db.QueryRow(ctx, getCompanyByIDSQL, id)
	var company Company
	err := row.Scan(&company.ID, &company.Name, &company.MonthlyTokenBudget, &company.CreatedAt, &company.UpdatedAt)
	return company, err
}

func (q *Queries) ListCompanies(ctx context.Context) ([]Company, error) {
	rows, err := q.db.Query(ctx, listCompaniesSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []Company
	for rows.Next() {
		var company Company
		if err := rows.Scan(&company.ID, &company.Name, &company.MonthlyTokenBudget, &company.CreatedAt, &company.UpdatedAt); err != nil {
			return nil, err
		}
		companies = append(companies, company)
	}
	return companies, rows.Err()
}

func (q *Queries) UpdateCompany(ctx context.Context, arg Company) error {
	_, err := q.db.Exec(ctx, updateCompanySQL, arg.ID, arg.Name, arg.MonthlyTokenBudget, arg.UpdatedAt)
	return err
}

func (q *Queries) DeleteCompany(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, deleteCompanySQL, id)
	return err
}
