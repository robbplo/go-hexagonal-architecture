-- name: CreateCompany :exec
INSERT INTO companies (id, name, monthly_token_budget, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetCompanyByID :one
SELECT id, name, monthly_token_budget, created_at, updated_at
FROM companies
WHERE id = $1;

-- name: ListCompanies :many
SELECT id, name, monthly_token_budget, created_at, updated_at
FROM companies
ORDER BY name ASC;

-- name: UpdateCompany :exec
UPDATE companies
SET name = $2, monthly_token_budget = $3, updated_at = $4
WHERE id = $1;

-- name: DeleteCompany :exec
DELETE FROM companies
WHERE id = $1;
