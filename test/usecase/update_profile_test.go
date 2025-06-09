package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pt-xyz-multifinance/internal/model"
	"github.com/pt-xyz-multifinance/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func TestUpdateProfile(t *testing.T) {
	mockRepo := new(MockUserRepo)
	useCase, err := usecase.NewUserUseCase(mockRepo, "test-secret", time.Hour)
	assert.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		existingUser := &model.User{
			ID:        "1",
			Username:  "testuser",
			Password:  "oldhash",
			CreatedAt: time.Now().Add(-24 * time.Hour),
		}

		updateUser := &model.User{
			ID:       "1",
			Username: "testuser",
			Email:    "newemail@example.com",
		}

		mockRepo.On("GetByID", mock.Anything, "1").Return(existingUser, nil)
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

		err := useCase.UpdateProfile(context.Background(), updateUser)

		assert.NoError(t, err)
		assert.Equal(t, existingUser.Password, updateUser.Password)
		assert.Equal(t, existingUser.CreatedAt, updateUser.CreatedAt)
		assert.NotNil(t, updateUser.UpdatedAt)
		mockRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		updateUser := &model.User{ID: "999"}
		mockRepo.On("GetByID", mock.Anything, "999").Return(nil, nil)

		err := useCase.UpdateProfile(context.Background(), updateUser)

		assert.Equal(t, usecase.ErrUserNotFound, err)
		mockRepo.AssertExpectations(t)
	})
	t.Run("repository error", func(t *testing.T) {
		updateUser := &model.User{ID: "1"}
		mockRepo.On("GetByID", mock.Anything, "1").Return(&model.User{}, nil)
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.User")).Return(errors.New("database error"))

		err := useCase.UpdateProfile(context.Background(), updateUser)

		assert.Equal(t, errors.New("database error"), err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("missing user ID", func(t *testing.T) {
		updateUser := &model.User{
			Username: "testuser",
			Email:    "test@example.com",
		}

		err := useCase.UpdateProfile(context.Background(), updateUser)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID is required")
	})

	t.Run("invalid email format", func(t *testing.T) {
		existingUser := &model.User{
			ID:        "1",
			Username:  "testuser",
			Email:     "old@example.com",
			Password:  "oldhash",
			CreatedAt: time.Now().Add(-24 * time.Hour),
		}

		updateUser := &model.User{
			ID:    "1",
			Email: "invalidemail",
		}

		mockRepo.On("GetByID", mock.Anything, "1").Return(existingUser, nil)

		err := useCase.UpdateProfile(context.Background(), updateUser)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email format")
		mockRepo.AssertExpectations(t)
	})
	t.Run("weak new password", func(t *testing.T) {
		existingUser := &model.User{
			ID:        "1",
			Username:  "testuser",
			Password:  "oldhash",
			CreatedAt: time.Now().Add(-24 * time.Hour),
		}

		updateUser := &model.User{
			ID: "1", Password: "weak123", // Missing capital letter and special character
		}

		mockRepo.On("GetByID", mock.Anything, "1").Return(existingUser, nil)

		err := useCase.UpdateProfile(context.Background(), updateUser)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Password must be at least 8 characters long and contain at least one uppercase letter, lowercase letter, number, and special character")
		mockRepo.AssertExpectations(t)
	})
}
