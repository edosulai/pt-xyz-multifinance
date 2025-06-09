package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/edosulai/pt-xyz-multifinance/internal/model"
	"github.com/edosulai/pt-xyz-multifinance/pkg/database"
)

// LoanRepositoryImpl implements LoanRepository interface using native SQL
type LoanRepositoryImpl struct {
	db *database.DB
}

// NewLoanRepository creates a new loan repository instance
func NewLoanRepository(db *database.DB) LoanRepository {
	return &LoanRepositoryImpl{db: db}
}

// Create creates a new loan application
func (r *LoanRepositoryImpl) Create(ctx context.Context, loan *model.Loan) error {
	query := `
		INSERT INTO loans (
			user_id, amount, tenure_months, purpose, status, 
			monthly_payment, interest_rate, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	now := time.Now()
	loan.CreatedAt = now
	loan.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		loan.UserID, loan.Amount, loan.TenureMonths, loan.Purpose, loan.Status,
		loan.MonthlyPayment, loan.InterestRate, loan.CreatedAt, loan.UpdatedAt,
	).Scan(&loan.ID)

	if err != nil {
		return fmt.Errorf("failed to create loan: %v", err)
	}
	return nil
}

// GetByID retrieves a loan by its ID
func (r *LoanRepositoryImpl) GetByID(ctx context.Context, id string) (*model.Loan, error) {
	query := `
		SELECT 
			l.id, l.user_id, l.amount, l.tenure_months, l.purpose, l.status,
			l.monthly_payment, l.interest_rate, l.disbursed_amount, l.disbursed_at,
			l.created_at, l.updated_at
		FROM loans l
		WHERE l.id = $1 AND l.deleted_at IS NULL`
	loan := &model.Loan{}
	var nullDisbursedAmount sql.NullFloat64
	var nullDisbursedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&loan.ID, &loan.UserID, &loan.Amount, &loan.TenureMonths, &loan.Purpose, &loan.Status,
		&loan.MonthlyPayment, &loan.InterestRate, &nullDisbursedAmount,
		&nullDisbursedAt, &loan.CreatedAt, &loan.UpdatedAt,
	)

	if nullDisbursedAmount.Valid {
		loan.DisbursedAmount = nullDisbursedAmount.Float64
	}
	if nullDisbursedAt.Valid {
		loan.DisbursedAt = &nullDisbursedAt.Time
	}

	if err == sql.ErrNoRows {
		return nil, errors.New("loan not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get loan: %v", err)
	}

	// Get documents
	documents, err := r.GetDocumentsByLoanID(ctx, id)
	if err != nil {
		return nil, err
	}
	loan.Documents = documents

	return loan, nil
}

// GetUserLoans retrieves all loans for a user with pagination
func (r *LoanRepositoryImpl) GetUserLoans(ctx context.Context, userID string, page, pageSize int) ([]model.Loan, int64, error) {
	var total int64
	countQuery := `
		SELECT COUNT(*) 
		FROM loans 
		WHERE user_id = $1 AND deleted_at IS NULL`

	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count loans: %v", err)
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT 
			l.id, l.user_id, l.amount, l.tenure_months, l.purpose, l.status,
			l.monthly_payment, l.interest_rate, l.disbursed_amount, l.disbursed_at,
			l.created_at, l.updated_at
		FROM loans l
		WHERE l.user_id = $1 AND l.deleted_at IS NULL
		ORDER BY l.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get loans: %v", err)
	}
	defer rows.Close()

	var loans []model.Loan
	for rows.Next() {
		var loan model.Loan
		err := rows.Scan(
			&loan.ID, &loan.UserID, &loan.Amount, &loan.TenureMonths, &loan.Purpose, &loan.Status,
			&loan.MonthlyPayment, &loan.InterestRate, &loan.DisbursedAmount,
			&loan.DisbursedAt, &loan.CreatedAt, &loan.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan loan: %v", err)
		}

		// Get documents for each loan
		documents, err := r.GetDocumentsByLoanID(ctx, loan.ID)
		if err != nil {
			return nil, 0, err
		}
		loan.Documents = documents

		loans = append(loans, loan)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating loans: %v", err)
	}

	return loans, total, nil
}

// UpdateLoanStatus updates the status of a loan
func (r *LoanRepositoryImpl) UpdateLoanStatus(ctx context.Context, id string, status model.LoanStatus) error {
	query := `
		UPDATE loans 
		SET status = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update loan status: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %v", err)
	}
	if rows == 0 {
		return fmt.Errorf("loan not found")
	}

	return nil
}

// UpdateLoan updates a loan record
func (r *LoanRepositoryImpl) UpdateLoan(ctx context.Context, loan *model.Loan) error {
	query := `
		UPDATE loans 
		SET user_id = $1, amount = $2, tenure_months = $3, purpose = $4,
			status = $5, monthly_payment = $6, interest_rate = $7,
			disbursed_amount = $8, disbursed_at = $9, updated_at = $10
		WHERE id = $11 AND deleted_at IS NULL`

	loan.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		loan.UserID, loan.Amount, loan.TenureMonths, loan.Purpose,
		loan.Status, loan.MonthlyPayment, loan.InterestRate,
		loan.DisbursedAmount, loan.DisbursedAt, loan.UpdatedAt,
		loan.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update loan: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %v", err)
	}
	if rows == 0 {
		return fmt.Errorf("loan not found")
	}

	return nil
}

// AddDocument adds a document to a loan
func (r *LoanRepositoryImpl) AddDocument(ctx context.Context, doc *model.Document) error {
	query := `
		INSERT INTO documents (
			loan_id, type, name, status, url, uploaded_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	now := time.Now()
	doc.CreatedAt = now
	doc.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		doc.LoanID, doc.Type, doc.Name, doc.Status, doc.URL,
		doc.UploadedAt, doc.CreatedAt, doc.UpdatedAt,
	).Scan(&doc.ID)

	if err != nil {
		return fmt.Errorf("failed to add document: %v", err)
	}
	return nil
}

// UpdateDocumentStatus updates the status of a document
func (r *LoanRepositoryImpl) UpdateDocumentStatus(ctx context.Context, id string, status model.DocumentStatus) error {
	query := `
		UPDATE documents 
		SET status = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update document status: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %v", err)
	}
	if rows == 0 {
		return fmt.Errorf("document not found")
	}

	return nil
}

// GetDocumentByID retrieves a document by its ID
func (r *LoanRepositoryImpl) GetDocumentByID(ctx context.Context, id string) (*model.Document, error) {
	query := `
		SELECT 
			id, loan_id, type, name, status, url, uploaded_at,
			created_at, updated_at
		FROM documents
		WHERE id = $1 AND deleted_at IS NULL`

	doc := &model.Document{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&doc.ID, &doc.LoanID, &doc.Type, &doc.Name, &doc.Status,
		&doc.URL, &doc.UploadedAt, &doc.CreatedAt, &doc.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %v", err)
	}

	return doc, nil
}

// GetDocumentsByLoanID retrieves all documents for a loan
func (r *LoanRepositoryImpl) GetDocumentsByLoanID(ctx context.Context, loanID string) ([]model.Document, error) {
	query := `
		SELECT 
			id, loan_id, type, name, status, url, uploaded_at,
			created_at, updated_at
		FROM documents
		WHERE loan_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, loanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents: %v", err)
	}
	defer rows.Close()

	var documents []model.Document
	for rows.Next() {
		var doc model.Document
		err := rows.Scan(
			&doc.ID, &doc.LoanID, &doc.Type, &doc.Name, &doc.Status,
			&doc.URL, &doc.UploadedAt, &doc.CreatedAt, &doc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document: %v", err)
		}
		documents = append(documents, doc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating documents: %v", err)
	}

	return documents, nil
}
