package usecase_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/pt-xyz-multifinance/internal/model"
	"github.com/pt-xyz-multifinance/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestRateLimit(t *testing.T) {
	mockRepo := new(MockUserRepo)
	useCase, err := usecase.NewUserUseCase(mockRepo, "test-secret", time.Hour)
	assert.NoError(t, err)

	// Create a test user
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("Test@123"), bcrypt.DefaultCost)
	testUser := &model.User{
		ID:       "1",
		Username: "testuser",
		Password: string(hashedPwd),
	}

	// Configure mock to return test user
	mockRepo.On("GetByUsername", mock.Anything, "testuser").Return(testUser, nil).Times(6)

	// Make login attempts
	var rateLimitHit bool
	for i := 0; i < 6; i++ {
		token, refresh, user, err := useCase.Login(context.Background(), "testuser", "wrongpassword")

		// All attempts should fail
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Empty(t, refresh)
		assert.Nil(t, user)

		if strings.Contains(err.Error(), "rate limit exceeded") {
			rateLimitHit = true
		} else {
			assert.Equal(t, usecase.ErrInvalidCredentials, err)
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Should have hit rate limit after multiple attempts
	assert.True(t, rateLimitHit, "Rate limit was never triggered")
	mockRepo.AssertExpectations(t)
}
