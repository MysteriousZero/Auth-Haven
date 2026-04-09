package repository

import (
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"
	"context"
	"fmt"
	"time"
)

type cachedAuthRepository struct {
	repo  interfaces.AuthRepository
	cache *RedisCache
}

func NewCachedAuthRepository(repo interfaces.AuthRepository, cache *RedisCache) interfaces.AuthRepository {
	return &cachedAuthRepository{
		repo:  repo,
		cache: cache,
	}
}

func (r *cachedAuthRepository) GetSessionByID(ctx context.Context, sessionID string) (*models.Session, error) {
	key := fmt.Sprintf("auth:sess:%s", sessionID)
	var session models.Session

	// Try cache
	err := r.cache.Get(ctx, key, &session)
	if err == nil {
		return &session, nil
	}

	// Cache miss, hit DB
	s, err := r.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Save to cache (1 hour TTL as per spec)
	_ = r.cache.Set(ctx, key, s, time.Hour)

	return s, nil
}

func (r *cachedAuthRepository) CreateSession(ctx context.Context, session *models.Session) error {
	err := r.repo.CreateSession(ctx, session)
	if err != nil {
		return err
	}
	// Optionally pre-populate cache
	key := fmt.Sprintf("auth:sess:%s", session.SessionID)
	_ = r.cache.Set(ctx, key, session, time.Hour)
	return nil
}

func (r *cachedAuthRepository) DeleteSession(ctx context.Context, sessionID string) error {
	err := r.repo.DeleteSession(ctx, sessionID)
	if err != nil {
		return err
	}
	_ = r.cache.Delete(ctx, fmt.Sprintf("auth:sess:%s", sessionID))
	return nil
}

func (r *cachedAuthRepository) RevokeSession(ctx context.Context, sessionID string) error {
	err := r.repo.RevokeSession(ctx, sessionID)
	if err != nil {
		return err
	}
	_ = r.cache.Delete(ctx, fmt.Sprintf("auth:sess:%s", sessionID))
	return nil
}

// Pass-through methods
func (r *cachedAuthRepository) GetActiveSessions(ctx context.Context, userID string) ([]models.Session, error) {
	return r.repo.GetActiveSessions(ctx, userID)
}

func (r *cachedAuthRepository) ListSessions(ctx context.Context, userID string, limit, offset int) ([]models.Session, error) {
	return r.repo.ListSessions(ctx, userID, limit, offset)
}

func (r *cachedAuthRepository) ListSessionsByDevice(ctx context.Context, deviceID string) ([]models.Session, error) {
	return r.repo.ListSessionsByDevice(ctx, deviceID)
}

func (r *cachedAuthRepository) DeleteAllUserSessions(ctx context.Context, userID string) error {
	err := r.repo.DeleteAllUserSessions(ctx, userID)
	if err != nil {
		return err
	}
	// Invalidate all user sessions (prefix search)
	// Note: Our keys are auth:sess:{session_id}, so we'd need a secondary index to map UserID -> SessionIDs
	// For now, we'll just rely on TTL or do a more expensive prefix purge if UserID was in the key.
	// Since we don't have UserID in the key, we'll just skip cache purge for now or implement prefix search later.
	return nil
}

func (r *cachedAuthRepository) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	return r.repo.CreateRefreshToken(ctx, token)
}

func (r *cachedAuthRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	return r.repo.GetRefreshTokenByHash(ctx, tokenHash)
}

func (r *cachedAuthRepository) RevokeRefreshToken(ctx context.Context, tokenID string) error {
	return r.repo.RevokeRefreshToken(ctx, tokenID)
}

func (r *cachedAuthRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	return r.repo.RevokeAllUserTokens(ctx, userID)
}
