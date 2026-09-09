package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"tugas1-go/pertemuan-5-authentication-security/app/model"
)

var ErrUserNotFound = errors.New("user tidak ditemukan")

type TokenRepository interface {
	Save(ctx context.Context, token model.RefreshToken) error
	FindActive(ctx context.Context, tokenHash string) (model.RefreshToken, error)
	Revoke(ctx context.Context, tokenHash string) error
	RevokeAllForUser(ctx context.Context, userID int) error
}

type tokenRepository struct {
	db *pgxpool.Pool
}

func NewTokenRepository(db *pgxpool.Pool) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) Save(ctx context.Context, token model.RefreshToken) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, token.UserID, token.TokenHash, token.ExpiresAt)
	if err != nil {
		return mapTokenDatabaseError("save refresh token", err)
	}
	return nil
}

func (r *tokenRepository) FindActive(ctx context.Context, tokenHash string) (model.RefreshToken, error) {
	var token model.RefreshToken
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > NOW()
	`, tokenHash).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt,
		&token.RevokedAt, &token.CreatedAt,
	)
	if err != nil {
		return model.RefreshToken{}, mapTokenDatabaseError("find active refresh token", err)
	}
	return token, nil
}

func (r *tokenRepository) Revoke(ctx context.Context, tokenHash string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1
		  AND revoked_at IS NULL
	`, tokenHash)
	if err != nil {
		return mapTokenDatabaseError("revoke refresh token", err)
	}
	return nil
}

func (r *tokenRepository) RevokeAllForUser(ctx context.Context, userID int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE user_id = $1
		  AND revoked_at IS NULL
	`, userID)
	if err != nil {
		return mapTokenDatabaseError("revoke all refresh tokens", err)
	}
	return nil
}

func mapTokenDatabaseError(operation string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			return ErrUserNotFound
		case "23505":
			return ErrDuplicate
		}
	}
	return fmt.Errorf("%s: %w", operation, err)
}
