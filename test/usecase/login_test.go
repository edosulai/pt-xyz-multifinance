package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/pt-xyz-multifinance/internal/model"
	"github.com/pt-xyz-multifinance/internal/repo"
	"github.com/pt-xyz-multifinance/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepo struct {
	mock.Mock
	repo.UserRepository
}

func TestLogin(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUserRepo)
	useCase, err := usecase.NewUserUseCase(mockRepo, "test-secret", time.Hour, usecase.WithoutRateLimiting())
	assert.NoError(t, err)

	password := "Test123!"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	t.Run("Successful login", func(t *testing.T) {
		user := &model.User{
			ID:       "1",
			Username: "testuser",
			Password: string(hashedPassword),
			Status:   "active",
		}
		mockRepo.On("GetByUsername", ctx, "testuser").Return(user, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*model.User")).Return(nil)

		token, refresh, loggedUser, err := useCase.Login(ctx, "testuser", password)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotEmpty(t, refresh)
		assert.Equal(t, user.ID, loggedUser.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid credentials", func(t *testing.T) {
		user := &model.User{
			ID:       "1",
			Username: "testuser",
			Password: string(hashedPassword),
			Status:   "active",
		}
		mockRepo.On("GetByUsername", ctx, "testuser").Return(user, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*model.User")).Return(nil)

		_, _, _, err := useCase.Login(ctx, "testuser", "wrongpassword")

		assert.Error(t, err)
		assert.Equal(t, usecase.ErrInvalidCredentials, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Account gets locked after 5 failed attempts", func(t *testing.T) {
		user := &model.User{
			ID:                  "1",
			Username:            "testuser",
			Password:            string(hashedPassword),
			Status:              "active",
			FailedLoginAttempts: 4,
		}
		mockRepo.On("GetByUsername", ctx, "testuser").Return(user, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(u *model.User) bool {
			return u.FailedLoginAttempts == 5 && u.Status == "suspended" && u.LockedUntil != nil
		})).Return(nil)

		_, _, _, err := useCase.Login(ctx, "testuser", "wrongpassword")

		assert.Error(t, err)
		assert.Equal(t, usecase.ErrInvalidCredentials, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Attempt to login to locked account", func(t *testing.T) {
		lockedUntil := time.Now().Add(15 * time.Minute)
		user := &model.User{
			ID:          "1",
			Username:    "testuser",
			Password:    string(hashedPassword),
			Status:      "suspended",
			LockedUntil: &lockedUntil,
		}

		mockRepo.On("GetByUsername", ctx, "testuser").Return(user, nil)

		_, _, _, err := useCase.Login(ctx, "testuser", password)

		assert.Error(t, err)
		assert.Equal(t, usecase.ErrAccountLocked, err)
		mockRepo.AssertExpectations(t)
	})
}
