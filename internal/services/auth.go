// Go
package services

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/maksemen2/avito-shop/internal/models"
	"github.com/maksemen2/avito-shop/internal/repository"
	"github.com/maksemen2/avito-shop/pkg/auth"
	"go.uber.org/zap"
)

type AuthService interface {
	Authenticate(ctx context.Context, req models.AuthRequest) (models.AuthResponse, error)
}

type authServiceImpl struct {
	repository repository.HolderRepository
	jwtManager *auth.JWTManager
	logger     *zap.Logger
}

func NewAuthService(repository repository.HolderRepository, jwtManager *auth.JWTManager, logger *zap.Logger) AuthService {
	return &authServiceImpl{
		repository: repository,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

func (s *authServiceImpl) Authenticate(ctx context.Context, req models.AuthRequest) (models.AuthResponse, error) {
	if req.Username == "" || req.Password == "" {
		return models.AuthResponse{}, ErrUserPassRequired
	}

	user, err := s.repository.User().GetByUsername(ctx, req.Username)
	wasCreated := false

	if err != nil {
		// Если пользователь не найден, регистрируем его
		if errors.Is(err, repository.ErrUserNotFound) {
			passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				s.logger.Error("Password hash generation failed", zap.Error(err))
				return models.AuthResponse{}, ErrInternal
			}

			user, err = s.repository.User().Create(ctx, req.Username, string(passwordHash))
			if err != nil {
				return models.AuthResponse{}, ErrInternal
			}

			wasCreated = true
		} else {
			return models.AuthResponse{}, ErrInternal
		}
	}

	// Сравнение хешей - довольно дорогая операция, поэтому производим её только если пользователь не был зарегестрирован
	// в процессе обработки этого запроса, так можно немного повысить производительность
	if !wasCreated {
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			return models.AuthResponse{}, ErrAuthFailed
		}
	}

	token, err := s.jwtManager.GenerateToken(user.ID, user.Username)
	if err != nil {
		s.logger.Error("Token generation failed", zap.Uint("userID", user.ID), zap.Error(err))
		return models.AuthResponse{}, ErrInternal
	}

	return models.AuthResponse{Token: token}, nil
}
