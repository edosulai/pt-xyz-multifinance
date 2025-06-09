package repo_test

import (
	"context"
	"testing"
	"time"

	"github.com/pt-xyz-multifinance/internal/model"
	"github.com/pt-xyz-multifinance/internal/repo"
	"github.com/pt-xyz-multifinance/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoanRepository(t *testing.T) {
	err := testutil.InitTestLogger(t)
	require.NoError(t, err, "Failed to initialize test logger")

	db := testutil.NewTestDB(t).DB
	loanRepo := repo.NewLoanRepository(db)
	userRepo := repo.NewUserRepository(db)

	// Create test user
	user := &model.User{
		Username:      "testuser",
		Email:         "test@example.com",
		Password:      "hashedpassword",
		FullName:      "Test User",
		PhoneNumber:   "+6281234567890",
		Address:       "Test Address",
		KTPNumber:     "1234567890123456",
		Status:        "active",
		MonthlyIncome: 5000000,
	}
	createErr := userRepo.Create(context.Background(), user)
	require.NoError(t, createErr, "Failed to create test user")

	t.Run("CreateLoan", func(t *testing.T) {
		loan := &model.Loan{
			UserID:       user.ID,
			Amount:       10000000,
			TenureMonths: 12,
			Purpose:      "Test purpose",
			Status:       model.LoanStatusPending,
			InterestRate: 10.5,
		}

		err := loanRepo.Create(context.Background(), loan)
		assert.NoError(t, err)
		assert.NotEmpty(t, loan.ID)
	})

	t.Run("GetByID", func(t *testing.T) {
		loan := &model.Loan{
			UserID:       user.ID,
			Amount:       15000000,
			TenureMonths: 24,
			Purpose:      "Another test",
			Status:       model.LoanStatusPending,
			InterestRate: 11.0,
		}

		err := loanRepo.Create(context.Background(), loan)
		require.NoError(t, err)
		found, err := loanRepo.GetByID(context.Background(), loan.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, loan.ID, found.ID)
		assert.Equal(t, loan.Amount, found.Amount)
	})

	t.Run("UpdateLoan", func(t *testing.T) {
		loan := &model.Loan{
			UserID:       user.ID,
			Amount:       20000000,
			TenureMonths: 36,
			Purpose:      "Test update",
			Status:       model.LoanStatusPending,
			InterestRate: 12.0,
		}

		err := loanRepo.Create(context.Background(), loan)
		require.NoError(t, err)

		loan.Status = model.LoanStatusApproved
		loan.MonthlyPayment = 650000

		err = loanRepo.UpdateLoan(context.Background(), loan)
		assert.NoError(t, err)
		updated, err := loanRepo.GetByID(context.Background(), loan.ID)
		assert.NoError(t, err)
		assert.Equal(t, model.LoanStatusApproved, updated.Status)
		assert.Equal(t, 650000.0, updated.MonthlyPayment)
	})

	t.Run("GetUserLoans", func(t *testing.T) {
		// Create multiple loans for user
		for i := 0; i < 5; i++ {
			loan := &model.Loan{
				UserID:       user.ID,
				Amount:       1000000 * float64(i+1),
				TenureMonths: 12,
				Purpose:      "Test pagination",
				Status:       model.LoanStatusPending,
				InterestRate: 10.0,
			}
			err := loanRepo.Create(context.Background(), loan)
			require.NoError(t, err)
		}

		loans, total, err := loanRepo.GetUserLoans(context.Background(), user.ID, 1, 3)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(5))
		assert.Len(t, loans, 3)
	})

	t.Run("AddDocument", func(t *testing.T) {
		loan := &model.Loan{
			UserID:       user.ID,
			Amount:       25000000,
			TenureMonths: 24,
			Purpose:      "Test with document",
			Status:       model.LoanStatusPending,
			InterestRate: 10.5,
		}

		err := loanRepo.Create(context.Background(), loan)
		require.NoError(t, err)

		doc := &model.Document{
			LoanID:     loan.ID,
			Type:       model.DocumentTypeKTP,
			Name:       "ktp.jpg",
			Status:     model.DocumentStatusUploaded,
			URL:        "http://storage.example.com/ktp.jpg",
			UploadedAt: &time.Time{},
		}

		err = loanRepo.AddDocument(context.Background(), doc)
		assert.NoError(t, err)
		assert.NotEmpty(t, doc.ID)

		// Verify document is linked to loan
		foundLoan, err := loanRepo.GetByID(context.Background(), loan.ID)
		assert.NoError(t, err)
		assert.Len(t, foundLoan.Documents, 1)
		assert.Equal(t, doc.ID, foundLoan.Documents[0].ID)
	})
}
