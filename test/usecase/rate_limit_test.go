package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/edosulai/pt-xyz-multifinance/internal/model"
	"github.com/edosulai/pt-xyz-multifinance/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestRateLimitedLogin(t *testing.T) {
	mockRepo := new(MockUserRepo)
	useCase, err := usecase.NewUserUseCase(mockRepo, "test-secret", time.Hour)
	assert.NoError(t, err)

	// Create a test user with hashed password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Test@123"), bcrypt.DefaultCost)
	testUser := &model.User{
		ID:       "1",
		Username: "testuser",
		Password: string(hashedPassword),
	}

	// Set up the mock to return our test user
	mockRepo.On("GetByUsername", mock.Anything, "testuser").Return(testUser, nil)

	// Try multiple login attempts
	for i := 0; i < 7; i++ {
		token, refreshToken, user, err := useCase.Login(context.Background(), "testuser", "wrongpassword")

		// First 5 attempts should fail with invalid credentials
		if i < 5 {
			assert.Equal(t, usecase.ErrInvalidCredentials, err)
		} else {
			// After 5 attempts, it should fail with rate limit error
			assert.Contains(t, err.Error(), "rate limit exceeded")
		}

		// All failed attempts should return no tokens or user
		assert.Empty(t, token)
		assert.Empty(t, refreshToken)
		assert.Nil(t, user)

		// Sleep a bit between attempts to ensure rate limit counting works properly
		time.Sleep(100 * time.Millisecond)
	}

	// Verify all mock expectations were met
	mockRepo.AssertExpectations(t)
}
