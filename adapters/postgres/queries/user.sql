-- name: InsertUserProfile :exec
INSERT INTO user_profiles (
    id, email, role, status, company_id, invited_at, activated_at, disabled_at, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: GetUserProfileByID :one
SELECT id, email, role, status, company_id, invited_at, activated_at, disabled_at, created_at, updated_at
FROM user_profiles
WHERE id = $1;

-- name: GetUserProfileByEmail :one
SELECT id, email, role, status, company_id, invited_at, activated_at, disabled_at, created_at, updated_at
FROM user_profiles
WHERE email = $1;

-- name: UpdateUserProfile :exec
UPDATE user_profiles
SET email = $2, role = $3, status = $4, invited_at = $5, activated_at = $6, disabled_at = $7, updated_at = $8
WHERE id = $1;

-- name: DeleteUserProfile :exec
DELETE FROM user_profiles
WHERE id = $1;
