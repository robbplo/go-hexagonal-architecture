package postgres

import "context"

const insertUserProfileSQL = `
INSERT INTO user_profiles (
	id, email, role, status, company_id, invited_at, activated_at, disabled_at, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`

const getUserProfileByIDSQL = `
SELECT id, email, role, status, company_id, invited_at, activated_at, disabled_at, created_at, updated_at
FROM user_profiles
WHERE id = $1
`

const getUserProfileByEmailSQL = `
SELECT id, email, role, status, company_id, invited_at, activated_at, disabled_at, created_at, updated_at
FROM user_profiles
WHERE email = $1
`

const updateUserProfileSQL = `
UPDATE user_profiles
SET email = $2, role = $3, status = $4, invited_at = $5, activated_at = $6, disabled_at = $7, updated_at = $8
WHERE id = $1
`

const deleteUserProfileSQL = `
DELETE FROM user_profiles
WHERE id = $1
`

func (q *Queries) InsertUserProfile(ctx context.Context, arg UserProfile) error {
	_, err := q.db.Exec(ctx, insertUserProfileSQL,
		arg.ID,
		arg.Email,
		arg.Role,
		arg.Status,
		arg.CompanyID,
		arg.InvitedAt,
		arg.ActivatedAt,
		arg.DisabledAt,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	return err
}

func (q *Queries) GetUserProfileByID(ctx context.Context, id string) (UserProfile, error) {
	row := q.db.QueryRow(ctx, getUserProfileByIDSQL, id)
	var user UserProfile
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Role,
		&user.Status,
		&user.CompanyID,
		&user.InvitedAt,
		&user.ActivatedAt,
		&user.DisabledAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	return user, err
}

func (q *Queries) GetUserProfileByEmail(ctx context.Context, email string) (UserProfile, error) {
	row := q.db.QueryRow(ctx, getUserProfileByEmailSQL, email)
	var user UserProfile
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Role,
		&user.Status,
		&user.CompanyID,
		&user.InvitedAt,
		&user.ActivatedAt,
		&user.DisabledAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	return user, err
}

func (q *Queries) UpdateUserProfile(ctx context.Context, arg UserProfile) error {
	_, err := q.db.Exec(ctx, updateUserProfileSQL,
		arg.ID,
		arg.Email,
		arg.Role,
		arg.Status,
		arg.InvitedAt,
		arg.ActivatedAt,
		arg.DisabledAt,
		arg.UpdatedAt,
	)
	return err
}

func (q *Queries) DeleteUserProfile(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, deleteUserProfileSQL, id)
	return err
}
