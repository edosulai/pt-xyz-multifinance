package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/pt-xyz-multifinance/internal/model"
	"github.com/pt-xyz-multifinance/internal/repo"
)

// LoanUseCase defines the interface for loan business logic
type LoanUseCase interface {
	// Apply for a new loan
	ApplyLoan(ctx context.Context, userID string, amount float64, tenureMonths int, purpose string) (*model.Loan, error)

	// Get loan application status
	GetLoanStatus(ctx context.Context, loanID string) (*model.Loan, error) // Get user's loan history
	GetLoanHistory(ctx context.Context, userID string, page, pageSize int) ([]model.Loan, int64, error)

	// Submit loan documents
	SubmitLoanDocuments(ctx context.Context, loanID string, docs []model.Document) error

	// Process loan application (for admin/system)
	ProcessLoanApplication(ctx context.Context, loanID string, approve bool, interestRate float64) error

	// Disburse approved loan (for admin/system)
	DisburseLoan(ctx context.Context, loanID string, disbursedAmount float64) error
}

// LoanUseCaseImpl implements LoanUseCase interface
type LoanUseCaseImpl struct {
	loanRepo repo.LoanRepository
	userRepo repo.UserRepository
}

// NewLoanUseCase creates a new loan use case instance
func NewLoanUseCase(loanRepo repo.LoanRepository, userRepo repo.UserRepository) LoanUseCase {
	return &LoanUseCaseImpl{
		loanRepo: loanRepo,
		userRepo: userRepo,
	}
}

// ApplyLoan handles the loan application process
func (uc *LoanUseCaseImpl) ApplyLoan(ctx context.Context, userID string, amount float64, tenureMonths int, purpose string) (*model.Loan, error) {
	// Validate user exists and is eligible
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Validate loan amount and tenure
	if amount < 1000000 {
		return nil, fmt.Errorf("loan amount must be at least 1,000,000")
	}
	if tenureMonths < 6 || tenureMonths > 60 {
		return nil, fmt.Errorf("loan tenure must be between 6 and 60 months")
	}

	// Create loan application
	loan := &model.Loan{
		UserID:       userID,
		Amount:       amount,
		TenureMonths: tenureMonths,
		Purpose:      purpose,
		Status:       model.LoanStatusPending,
	}

	if err := uc.loanRepo.Create(ctx, loan); err != nil {
		return nil, err
	}

	return loan, nil
}

// GetLoanStatus retrieves the current status of a loan application
func (uc *LoanUseCaseImpl) GetLoanStatus(ctx context.Context, loanID string) (*model.Loan, error) {
	return uc.loanRepo.GetByID(ctx, loanID)
}

// GetLoanHistory retrieves the loan history for a user
func (uc *LoanUseCaseImpl) GetLoanHistory(ctx context.Context, userID string, page, pageSize int) ([]model.Loan, int64, error) {
	return uc.loanRepo.GetUserLoans(ctx, userID, page, pageSize)
}

// SubmitLoanDocuments handles document submission for a loan
func (uc *LoanUseCaseImpl) SubmitLoanDocuments(ctx context.Context, loanID string, docs []model.Document) error {
	loan, err := uc.loanRepo.GetByID(ctx, loanID)
	if err != nil {
		return err
	}
	if loan == nil {
		return ErrLoanNotFound
	}

	// Validate loan status
	if loan.Status != model.LoanStatusPending && loan.Status != model.LoanStatusInReview {
		return fmt.Errorf("cannot submit documents for loan in %s status", loan.Status)
	}

	// Add documents
	for _, doc := range docs {
		doc.LoanID = loanID
		doc.Status = model.DocumentStatusUploaded
		doc.UploadedAt = timePtr(time.Now())
		if err := uc.loanRepo.AddDocument(ctx, &doc); err != nil {
			return err
		}
	}

	// Update loan status to in_review
	return uc.loanRepo.UpdateLoanStatus(ctx, loanID, model.LoanStatusInReview)
}

// ProcessLoanApplication handles the loan approval/rejection process
func (uc *LoanUseCaseImpl) ProcessLoanApplication(ctx context.Context, loanID string, approve bool, interestRate float64) error {
	loan, err := uc.loanRepo.GetByID(ctx, loanID)
	if err != nil {
		return err
	}
	if loan == nil {
		return ErrLoanNotFound
	}

	if approve {
		if interestRate <= 0 {
			return fmt.Errorf("interest rate must be greater than 0")
		}
		loan.Status = model.LoanStatusApproved
		loan.InterestRate = interestRate
		loan.MonthlyPayment = calculateMonthlyPayment(loan.Amount, interestRate, loan.TenureMonths)
	} else {
		loan.Status = model.LoanStatusRejected
	}

	return uc.loanRepo.UpdateLoan(ctx, loan)
}

// DisburseLoan handles the loan disbursement process
func (uc *LoanUseCaseImpl) DisburseLoan(ctx context.Context, loanID string, disbursedAmount float64) error {
	loan, err := uc.loanRepo.GetByID(ctx, loanID)
	if err != nil {
		return err
	}
	if loan == nil {
		return ErrLoanNotFound
	}

	if loan.Status != model.LoanStatusApproved {
		return fmt.Errorf("cannot disburse loan in %s status", loan.Status)
	}

	if disbursedAmount <= 0 || disbursedAmount > loan.Amount {
		return fmt.Errorf("invalid disbursement amount")
	}

	now := time.Now()
	loan.Status = model.LoanStatusDisbursed
	loan.DisbursedAmount = disbursedAmount
	loan.DisbursedAt = &now

	return uc.loanRepo.UpdateLoan(ctx, loan)
}

// Helper function to create Time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}

// Helper function to calculate monthly payment
func calculateMonthlyPayment(principal, interestRate float64, tenureMonths int) float64 {
	monthlyRate := interestRate / 12 / 100

	// Using the PMT formula: PMT = P * (r * (1 + r)^n) / ((1 + r)^n - 1)
	temp := (1 + monthlyRate)
	factor := temp
	for i := 1; i < tenureMonths; i++ {
		factor *= temp
	}

	return principal * (monthlyRate * factor) / (factor - 1)
}
