package auth

const (
	queryGetUserByEmail = `
		SELECT id, email, display_name, password_hash, role, is_active, created_at, updated_at, last_login_at
		FROM woulder.users
		WHERE lower(email) = lower($1)
	`
	queryGetUserByID = `
		SELECT id, email, display_name, password_hash, role, is_active, created_at, updated_at, last_login_at
		FROM woulder.users
		WHERE id = $1
	`
	queryCreateUser = `
		INSERT INTO woulder.users (email, display_name, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, display_name, password_hash, role, is_active, created_at, updated_at, last_login_at
	`
	queryUpdateLastLogin    = `UPDATE woulder.users SET last_login_at = now(), updated_at = now() WHERE id = $1`
	queryCreateRefreshToken = `
		INSERT INTO woulder.auth_refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, token_hash, expires_at, revoked_at, created_at
	`
	queryGetRefreshToken = `
		SELECT rt.id, rt.user_id, rt.token_hash, rt.expires_at, rt.revoked_at, rt.created_at,
		       u.id, u.email, u.display_name, u.password_hash, u.role, u.is_active, u.created_at, u.updated_at, u.last_login_at
		FROM woulder.auth_refresh_tokens rt
		JOIN woulder.users u ON u.id = rt.user_id
		WHERE rt.token_hash = $1
	`
	queryRevokeRefreshToken = `UPDATE woulder.auth_refresh_tokens SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL`
)
