package repo_test

import (
	"context"
	"testing"
	"time"

	"github.com/edosulai/pt-xyz-multifinance/internal/model"
	"github.com/edosulai/pt-xyz-multifinance/internal/repo"
	"github.com/edosulai/pt-xyz-multifinance/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.NewTestDB(t)
	defer testDB.Cleanup()

	userRepo := repo.NewUserRepository(testDB.DB)

	t.Run("Create and Get User", func(t *testing.T) {
		// Clean up before test
		err := testDB.TruncateTables(ctx)
		require.NoError(t, err)

		// Create test user
		user := &model.User{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "hashedpassword",
			FullName: "Test User",
		}

		err = userRepo.Create(ctx, user)
		require.NoError(t, err)
		assert.NotEmpty(t, user.ID)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())

		// Get by ID
		found, err := userRepo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, user.Username, found.Username)
		assert.Equal(t, user.Email, found.Email)
		assert.Equal(t, user.FullName, found.FullName)

		// Get by username
		foundByUsername, err := userRepo.GetByUsername(ctx, user.Username)
		require.NoError(t, err)
		assert.NotNil(t, foundByUsername)
		assert.Equal(t, user.ID, foundByUsername.ID)

		// Get by email
		foundByEmail, err := userRepo.GetByEmail(ctx, user.Email)
		require.NoError(t, err)
		assert.NotNil(t, foundByEmail)
		assert.Equal(t, user.ID, foundByEmail.ID)
	})

	t.Run("Update User", func(t *testing.T) {
		// Clean up before test
		err := testDB.TruncateTables(ctx)
		require.NoError(t, err)

		// Create test user
		user := &model.User{
			Username: "updateuser",
			Email:    "update@example.com",
			Password: "hashedpassword",
			FullName: "Update User",
		}

		err = userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Update user
		user.FullName = "Updated Name"
		time.Sleep(time.Second) // Ensure updated_at will be different
		err = userRepo.Update(ctx, user)
		require.NoError(t, err)

		// Verify update
		found, err := userRepo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "Updated Name", found.FullName)
		assert.True(t, found.UpdatedAt.After(found.CreatedAt))
	})

	t.Run("Delete User", func(t *testing.T) {
		// Clean up before test
		err := testDB.TruncateTables(ctx)
		require.NoError(t, err)

		// Create test user
		user := &model.User{
			Username: "deleteuser",
			Email:    "delete@example.com",
			Password: "hashedpassword",
			FullName: "Delete User",
		}

		err = userRepo.Create(ctx, user)
		require.NoError(t, err)

		// Delete user
		err = userRepo.Delete(ctx, user.ID)
		require.NoError(t, err)

		// Verify deletion
		found, err := userRepo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Nil(t, found)
	})
}
