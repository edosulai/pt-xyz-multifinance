package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents the user entity in the database
type User struct {
	ID                  string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Username            string         `gorm:"uniqueIndex;not null" json:"username" validate:"required,min=3,max=50"`
	Email               string         `gorm:"uniqueIndex;not null" json:"email" validate:"required,email"`
	Password            string         `gorm:"not null" json:"-" validate:"required,min=8"`
	FullName            string         `gorm:"not null" json:"full_name" validate:"required"`
	PhoneNumber         string         `gorm:"unique;not null" json:"phone_number" validate:"required,e164"`
	Address             string         `gorm:"type:text;not null" json:"address" validate:"required"`
	KTPNumber           string         `gorm:"unique;not null" json:"ktp_number" validate:"required,len=16"`
	Status              string         `gorm:"not null;default:'active'" json:"status" validate:"required,oneof=active inactive suspended"`
	MonthlyIncome       float64        `gorm:"type:decimal(15,2);not null" json:"monthly_income" validate:"required,min=0"`
	FailedLoginAttempts int            `gorm:"default:0" json:"failed_login_attempts"`
	LastFailedLogin     *time.Time     `json:"last_failed_login,omitempty"`
	LockedUntil         *time.Time     `json:"locked_until,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
	Loans               []Loan         `gorm:"foreignKey:UserID" json:"loans,omitempty"`
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate is a hook that runs before creating a new user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.CreatedAt = time.Now()
	}
	u.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate is a hook that runs before updating a user
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}
