package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/edosulai/pt-xyz-multifinance/internal/model"
	"github.com/edosulai/pt-xyz-multifinance/pkg/database"
	"github.com/lib/pq"
)

// UserRepositoryImpl implements LoanRepository interface using native SQL
type UserRepositoryImpl struct {
	db *database.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *database.DB) UserRepository {
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (
			username, email, password, full_name, phone_number,
			address, ktp_number, status, monthly_income,
			failed_login_attempts, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		) RETURNING id`

	now := time.Now()
	err := r.db.QueryRowContext(ctx, query,
		user.Username, user.Email, user.Password, user.FullName,
		user.PhoneNumber, user.Address, user.KTPNumber, user.Status,
		user.MonthlyIncome, 0, now, now,
	).Scan(&user.ID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				if pqErr.Constraint == "users_username_key" {
					return errors.New("username already exists")
				}
				if pqErr.Constraint == "users_email_key" {
					return errors.New("email already exists")
				}
			}
		}
		return err
	}
	return nil
}

func (r *UserRepositoryImpl) GetByID(ctx context.Context, id string) (*model.User, error) {
	query := `
		SELECT id, username, email, password, full_name, phone_number,
			   address, ktp_number, status, monthly_income,
			   failed_login_attempts, last_failed_login, locked_until,
			   created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL`
	user, err := r.scanSingleUser(ctx, query, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found") // Keep this error for GetByID since it's expected to find a user
	}
	return user, nil
}

func (r *UserRepositoryImpl) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
		SELECT id, username, email, password, full_name, phone_number,
			   address, ktp_number, status, monthly_income,
			   failed_login_attempts, last_failed_login, locked_until,
			   created_at, updated_at
		FROM users
		WHERE username = $1 AND deleted_at IS NULL`

	user, err := r.scanSingleUser(ctx, query, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, username, email, password, full_name, phone_number,
			   address, ktp_number, status, monthly_income,
			   failed_login_attempts, last_failed_login, locked_until,
			   created_at, updated_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL`

	user, err := r.scanSingleUser(ctx, query, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *UserRepositoryImpl) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET username = $1,
			email = $2,
			password = $3,
			full_name = $4,
			phone_number = $5,
			address = $6,
			ktp_number = $7,
			status = $8,
			monthly_income = $9,
			failed_login_attempts = $10,
			last_failed_login = $11,
			locked_until = $12,
			updated_at = $13
		WHERE id = $14 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query,
		user.Username, user.Email, user.Password, user.FullName,
		user.PhoneNumber, user.Address, user.KTPNumber, user.Status,
		user.MonthlyIncome, user.FailedLoginAttempts,
		user.LastFailedLogin, user.LockedUntil,
		time.Now(), user.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *UserRepositoryImpl) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE users
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *UserRepositoryImpl) scanSingleUser(ctx context.Context, query string, args ...interface{}) (*model.User, error) {
	var user model.User
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.FullName, &user.PhoneNumber, &user.Address,
		&user.KTPNumber, &user.Status, &user.MonthlyIncome,
		&user.FailedLoginAttempts, &user.LastFailedLogin,
		&user.LockedUntil, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
