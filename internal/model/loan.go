package model

import (
	"time"

	"gorm.io/gorm"
)

// LoanStatus represents the status of a loan application
type LoanStatus string

const (
	LoanStatusPending      LoanStatus = "pending"
	LoanStatusInReview     LoanStatus = "in_review"
	LoanStatusApproved     LoanStatus = "approved"
	LoanStatusRejected     LoanStatus = "rejected"
	LoanStatusDisbursed    LoanStatus = "disbursed"
	LoanStatusPaidOff      LoanStatus = "paid_off"
	LoanStatusDefaulted    LoanStatus = "defaulted"
	LoanStatusRestructured LoanStatus = "restructured"
)

// DocumentStatus represents the status of a required document
type DocumentStatus string

const (
	DocumentStatusRequired DocumentStatus = "required"
	DocumentStatusUploaded DocumentStatus = "uploaded"
	DocumentStatusVerified DocumentStatus = "verified"
	DocumentStatusRejected DocumentStatus = "rejected"
)

// DocumentType represents the type of document required for loan
type DocumentType string

const (
	DocumentTypeKTP            DocumentType = "ktp"
	DocumentTypePayslip        DocumentType = "payslip"
	DocumentTypeBankStatement  DocumentType = "bank_statement"
	DocumentTypeEmployeeLetter DocumentType = "employee_letter"
)

// Loan represents a loan application in the system
type Loan struct {
	ID              string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID          string         `gorm:"not null" json:"user_id"`
	Amount          float64        `gorm:"type:decimal(15,2);not null" json:"amount" validate:"required,min=1000000"`
	TenureMonths    int            `gorm:"not null" json:"tenure_months" validate:"required,min=6,max=60"`
	Purpose         string         `gorm:"not null" json:"purpose" validate:"required"`
	Status          LoanStatus     `gorm:"not null;default:'pending'" json:"status"`
	MonthlyPayment  float64        `gorm:"type:decimal(15,2)" json:"monthly_payment"`
	InterestRate    float64        `gorm:"type:decimal(5,2);not null" json:"interest_rate"`
	DisbursedAmount float64        `gorm:"type:decimal(15,2)" json:"disbursed_amount"`
	DisbursedAt     *time.Time     `json:"disbursed_at"`
	Documents       []Document     `gorm:"foreignKey:LoanID" json:"documents"`
	User            User           `gorm:"foreignKey:UserID" json:"-"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// Document represents a document required for loan processing
type Document struct {
	ID         string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	LoanID     string         `gorm:"not null" json:"loan_id"`
	Type       DocumentType   `gorm:"not null" json:"type"`
	Name       string         `gorm:"not null" json:"name"`
	Status     DocumentStatus `gorm:"not null;default:'required'" json:"status"`
	URL        string         `gorm:"type:text" json:"url"`
	UploadedAt *time.Time     `json:"uploaded_at"`
	Loan       Loan           `gorm:"foreignKey:LoanID" json:"-"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate hook for Loan
func (l *Loan) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.CreatedAt = time.Now()
	}
	l.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook for Loan
func (l *Loan) BeforeUpdate(tx *gorm.DB) error {
	l.UpdatedAt = time.Now()
	return nil
}

// BeforeCreate hook for Document
func (d *Document) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.CreatedAt = time.Now()
	}
	d.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook for Document
func (d *Document) BeforeUpdate(tx *gorm.DB) error {
	d.UpdatedAt = time.Now()
	return nil
}
