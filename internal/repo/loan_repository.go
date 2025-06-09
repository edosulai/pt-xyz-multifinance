package repo

import (
	"context"

	"github.com/pt-xyz-multifinance/internal/model"
)

// LoanRepository defines the interface for loan data access
type LoanRepository interface { // Create a new loan application
	Create(ctx context.Context, loan *model.Loan) error
	// Get loan by ID
	GetByID(ctx context.Context, id string) (*model.Loan, error)

	// Get loans by user ID
	GetUserLoans(ctx context.Context, userID string, page, pageSize int) ([]model.Loan, int64, error)

	// Update loan status
	UpdateLoanStatus(ctx context.Context, id string, status model.LoanStatus) error

	// Update loan
	UpdateLoan(ctx context.Context, loan *model.Loan) error

	// Add document to loan
	AddDocument(ctx context.Context, doc *model.Document) error

	// Update document status
	UpdateDocumentStatus(ctx context.Context, id string, status model.DocumentStatus) error

	// Get document by ID
	GetDocumentByID(ctx context.Context, id string) (*model.Document, error)

	// Get documents by loan ID
	GetDocumentsByLoanID(ctx context.Context, loanID string) ([]model.Document, error)
}
