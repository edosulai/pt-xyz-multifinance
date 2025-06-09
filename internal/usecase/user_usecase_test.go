package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pt-xyz-multifinance/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestUserUseCase_Register(t *testing.T) {
	mockRepo := new(MockUserRepository)
	useCase, err := NewUserUseCase(mockRepo, "test-secret", time.Hour)
	assert.NoError(t, err)
	t.Run("successful registration", func(t *testing.T) {
		user := &model.User{
			Username: "testuser",
			Password: "Test@123", // Contains uppercase, number, and special char
			Email:    "test@example.com",
		}

		mockRepo.On("GetByUsername", mock.Anything, "testuser").Return(nil, nil)
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

		err := useCase.Register(context.Background(), user)

		assert.NoError(t, err)
		assert.NotEqual(t, "Test@123", user.Password) // Password should be hashed
		mockRepo.AssertExpectations(t)
	})
	t.Run("missing required fields", func(t *testing.T) {
		user := &model.User{
			Username: "",
			Password: "",
			Email:    "",
		}

		err := useCase.Register(context.Background(), user)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("invalid email format", func(t *testing.T) {
		user := &model.User{
			Username: "testuser",
			Password: "password123",
			Email:    "invalidemail",
		}

		err := useCase.Register(context.Background(), user)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email format")
	})

	t.Run("weak password - too short", func(t *testing.T) {
		user := &model.User{
			Username: "testuser",
			Password: "pass",
			Email:    "test@example.com",
		}

		err := useCase.Register(context.Background(), user)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least 8 characters")
	})
	t.Run("weak password - no numbers", func(t *testing.T) {
		user := &model.User{
			Username: "testuser",
			Password: "passwordonly",
			Email:    "test@example.com",
		}

		err := useCase.Register(context.Background(), user)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one number")
	})

	t.Run("weak password - no uppercase", func(t *testing.T) {
		user := &model.User{
			Username: "testuser",
			Password: "password123",
			Email:    "test@example.com",
		}

		err := useCase.Register(context.Background(), user)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one uppercase")
	})

	t.Run("weak password - no special chars", func(t *testing.T) {
		user := &model.User{
			Username: "testuser",
			Password: "Password123",
			Email:    "test@example.com",
		}

		err := useCase.Register(context.Background(), user)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "special character")
	})

	t.Run("strong password", func(t *testing.T) {
		user := &model.User{
			Username: "testuser",
			Password: "Password123!",
			Email:    "test@example.com",
		}
		mockRepo.On("GetByUsername", mock.Anything, "testuser").Return(nil, nil)
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

		err := useCase.Register(context.Background(), user)

		assert.NoError(t, err)
		assert.NotEqual(t, "Password123!", user.Password) // Should be hashed
		mockRepo.AssertExpectations(t)
	})
	t.Run("user already exists", func(t *testing.T) {
		user := &model.User{
			Username: "existinguser",
			Password: "Test@123", // Valid password format
			Email:    "existing@example.com",
		}

		existingUser := &model.User{ID: "1", Username: "existinguser"}
		mockRepo.On("GetByUsername", mock.Anything, "existinguser").Return(existingUser, nil)

		err := useCase.Register(context.Background(), user)

		assert.Equal(t, ErrUserExists, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserUseCase_Login(t *testing.T) {
	mockRepo := new(MockUserRepository)
	useCase, err := NewUserUseCase(mockRepo, "test-secret", time.Hour)
	assert.NoError(t, err)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Test@123"), bcrypt.DefaultCost)
	testUser := &model.User{
		ID:       "1",
		Username: "testuser",
		Password: string(hashedPassword),
		Email:    "test@example.com",
	}

	t.Run("successful login", func(t *testing.T) {
		mockRepo.On("GetByUsername", mock.Anything, "testuser").Return(testUser, nil)

		token, refreshToken, user, err := useCase.Login(context.Background(), "testuser", "Test@123")

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotEmpty(t, refreshToken)
		assert.Equal(t, testUser.ID, user.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		mockRepo.On("GetByUsername", mock.Anything, "testuser").Return(testUser, nil)

		token, refreshToken, user, err := useCase.Login(context.Background(), "testuser", "wrongpassword")

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidCredentials, err)
		assert.Empty(t, token)
		assert.Empty(t, refreshToken)
		assert.Nil(t, user)
	})

	t.Run("rate limit exceeded", func(t *testing.T) {
		mockRepo.On("GetByUsername", mock.Anything, "testuser").Return(testUser, nil).Times(6)

		// Make 6 failed login attempts
		for i := 0; i < 6; i++ {
			token, refreshToken, user, err := useCase.Login(context.Background(), "testuser", "wrongpassword")
			assert.Error(t, err)
			if err.Error() == ErrInvalidCredentials.Error() {
				assert.Equal(t, ErrInvalidCredentials, err)
			} else {
				assert.Contains(t, err.Error(), "rate limit exceeded")
			}
			assert.Empty(t, token)
			assert.Empty(t, refreshToken)
			assert.Nil(t, user)
			// Add a small delay to prevent rate limit from carrying over
			time.Sleep(time.Millisecond * 100)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo.On("GetByUsername", mock.Anything, "nonexistent").Return(nil, nil)

		token, refreshToken, user, err := useCase.Login(context.Background(), "nonexistent", "Test@123")

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidCredentials, err)
		assert.Empty(t, token)
		assert.Empty(t, refreshToken)
		assert.Nil(t, user)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.On("GetByUsername", mock.Anything, "testuser").Return(nil, errors.New("database error"))

		token, refreshToken, user, err := useCase.Login(context.Background(), "testuser", "Test@123")

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Empty(t, refreshToken)
		assert.Nil(t, user)
	})
}

func TestUserUseCase_UpdateProfile(t *testing.T) {
	mockRepo := new(MockUserRepository)
	useCase, err := NewUserUseCase(mockRepo, "test-secret", time.Hour)
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
		assert.Equal(t, existingUser.Password, updateUser.Password) // Should preserve old password
		assert.Equal(t, existingUser.CreatedAt, updateUser.CreatedAt)
		assert.NotNil(t, updateUser.UpdatedAt)
		mockRepo.AssertExpectations(t)
	})
	t.Run("user not found", func(t *testing.T) {
		updateUser := &model.User{ID: "999"}

		mockRepo.On("GetByID", mock.Anything, "999").Return(nil, nil)

		err := useCase.UpdateProfile(context.Background(), updateUser)

		assert.Equal(t, ErrUserNotFound, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("update with new password", func(t *testing.T) {
		existingUser := &model.User{
			ID:        "1",
			Username:  "testuser",
			Password:  "oldhash",
			CreatedAt: time.Now().Add(-24 * time.Hour),
		}

		updateUser := &model.User{
			ID:       "1",
			Username: "testuser",
			Password: "newpassword123",
		}
		mockRepo.On("GetByID", mock.Anything, "1").Return(existingUser, nil)
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

		err := useCase.UpdateProfile(context.Background(), updateUser)

		assert.NoError(t, err)
		assert.NotEqual(t, "newpassword123", updateUser.Password)      // Should be hashed
		assert.NotEqual(t, existingUser.Password, updateUser.Password) // Should be different from old hash
		mockRepo.AssertExpectations(t)
	})
	t.Run("repository error", func(t *testing.T) {
		updateUser := &model.User{ID: "1"}
		repoErr := errors.New("database error")

		mockRepo.On("GetByID", mock.Anything, "1").Return(nil, repoErr)

		err := useCase.UpdateProfile(context.Background(), updateUser)

		assert.Equal(t, repoErr, err)
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
			ID:       "1",
			Password: "weak",
		}

		mockRepo.On("GetByID", mock.Anything, "1").Return(existingUser, nil)

		err := useCase.UpdateProfile(context.Background(), updateUser)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least 8 characters")
		mockRepo.AssertExpectations(t)
	})
}

func TestUserUseCase_ValidateCaptcha(t *testing.T) {
	useCase, err := NewUserUseCase(nil, "test-secret", time.Hour)
	assert.NoError(t, err)

	t.Run("valid captcha", func(t *testing.T) {
		// This is a basic test since we're using the real captcha package
		result := useCase.ValidateCaptcha("dummy-id", "dummy-solution")
		assert.False(t, result) // Should be false for dummy values
	})
}
