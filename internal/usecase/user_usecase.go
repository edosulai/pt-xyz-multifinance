package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/dchest/captcha"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pt-xyz-multifinance/internal/model"
	"github.com/pt-xyz-multifinance/internal/repo"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"golang.org/x/crypto/bcrypt"
)

type UserUseCase interface {
	Register(ctx context.Context, user *model.User) error
	Login(ctx context.Context, username, password string) (string, string, *model.User, error)
	GetProfile(ctx context.Context, id string) (*model.User, error)
	UpdateProfile(ctx context.Context, user *model.User) error
	ValidateCaptcha(id, solution string) bool
}

type userUseCase struct {
	userRepo       repo.UserRepository
	jwtSecret      []byte
	jwtDuration    time.Duration
	loginLimiter   *limiter.Limiter
	disableLimiter bool // for testing
}

func NewUserUseCase(userRepo repo.UserRepository, jwtSecret string, jwtDuration time.Duration, opts ...UserOption) (UserUseCase, error) {
	uc := &userUseCase{
		userRepo:    userRepo,
		jwtSecret:   []byte(jwtSecret),
		jwtDuration: jwtDuration,
	}

	for _, opt := range opts {
		opt(uc)
	}

	if !uc.disableLimiter {
		// Initialize rate limiter: 5 attempts per minute per user
		rate, err := limiter.NewRateFromFormatted("5-M")
		if err != nil {
			return nil, fmt.Errorf("failed to create rate limiter: %w", err)
		}

		store := memory.NewStore()
		uc.loginLimiter = limiter.New(store, rate)
	}

	return uc, nil
}

// checkLoginRateLimit checks if the user has exceeded their login attempt limit
func (u *userUseCase) checkLoginRateLimit(ctx context.Context, identifier string) error {
	if u.disableLimiter {
		return nil
	}

	// Get the rate limit context for this user
	limiterCtx, err := u.loginLimiter.Get(ctx, identifier)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}

	if limiterCtx.Reached {
		resetTime := time.Unix(limiterCtx.Reset, 0)
		return fmt.Errorf("rate limit exceeded. Try again in %v", time.Until(resetTime))
	}

	return nil
}

func (u *userUseCase) Register(ctx context.Context, user *model.User) error {
	// Validate required fields
	if user.Username == "" || user.Password == "" || user.Email == "" {
		return errors.New("username, password and email are required")
	}

	// Validate email format
	if !strings.Contains(user.Email, "@") || !strings.Contains(user.Email, ".") {
		return errors.New("invalid email format")
	}
	// Validate password strength (minimum 8 characters, at least one number, uppercase, and special char)
	if len(user.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	var (
		hasNumber  bool
		hasUpper   bool
		hasSpecial bool
	)

	for _, char := range user.Password {
		switch {
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsUpper(char):
			hasUpper = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	if !hasNumber {
		return errors.New("password must contain at least one number")
	}
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	// Check if user already exists
	existingUser, err := u.userRepo.GetByUsername(ctx, user.Username)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	// Set initial status and timestamps
	user.Status = "active"
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	return u.userRepo.Create(ctx, user)
}

func (u *userUseCase) Login(ctx context.Context, username, password string) (string, string, *model.User, error) {
	// Check rate limit for this username
	if err := u.checkLoginRateLimit(ctx, username); err != nil {
		return "", "", nil, err
	}

	// Get user by username
	user, err := u.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", "", nil, err
	}
	if user == nil {
		// Update rate limit counter even for non-existent users
		_, _ = u.loginLimiter.Get(ctx, username)
		return "", "", nil, ErrInvalidCredentials
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// Update rate limit counter for failed attempts
		_, _ = u.loginLimiter.Get(ctx, username)
		return "", "", nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(u.jwtDuration).Unix(),
	})

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(u.jwtSecret)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to create token: %w", err)
	}

	// Generate refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(u.jwtDuration * 24).Unix(), // Refresh token valid for 24x longer
	})

	refreshTokenString, err := refreshToken.SignedString(u.jwtSecret)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return tokenString, refreshTokenString, user, nil
}

func (u *userUseCase) GetProfile(ctx context.Context, id string) (*model.User, error) {
	return u.userRepo.GetByID(ctx, id)
}

func (u *userUseCase) UpdateProfile(ctx context.Context, user *model.User) error {
	if user.ID == "" {
		return errors.New("user ID is required")
	}

	// Get existing user to verify it exists
	existingUser, err := u.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return err
	}
	if existingUser == nil {
		return ErrUserNotFound
	}

	// If email is being updated, validate format
	if user.Email != "" && user.Email != existingUser.Email {
		if !strings.Contains(user.Email, "@") || !strings.Contains(user.Email, ".") {
			return errors.New("invalid email format")
		}
	} else {
		user.Email = existingUser.Email
	}

	// If password is provided, validate and hash it
	if user.Password != "" {
		if len(user.Password) < 8 {
			return errors.New("password must be at least 8 characters long")
		}
		hasNumber := false
		for _, char := range user.Password {
			if unicode.IsNumber(char) {
				hasNumber = true
				break
			}
		}
		if !hasNumber {
			return errors.New("password must contain at least one number")
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
	} else {
		// Keep existing password if not updating
		user.Password = existingUser.Password
	}

	// Preserve certain fields from existing user
	user.CreatedAt = existingUser.CreatedAt
	user.UpdatedAt = time.Now()

	return u.userRepo.Update(ctx, user)
}

func (u *userUseCase) ValidateCaptcha(id, solution string) bool {
	// Validate the captcha solution using dchest/captcha package
	return captcha.VerifyString(id, solution)
}

func (u *userUseCase) generateTokens(user *model.User) (string, string, error) {
	// Generate access token
	claims := jwt.MapClaims{
		"user_id":    user.ID,
		"username":   user.Username,
		"exp":        time.Now().Add(u.jwtDuration).Unix(),
		"iat":        time.Now().Unix(),
		"token_type": "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(u.jwtSecret)
	if err != nil {
		return "", "", err
	}

	// Generate refresh token with longer expiration
	refreshClaims := jwt.MapClaims{
		"user_id":    user.ID,
		"exp":        time.Now().Add(u.jwtDuration * 24 * 7).Unix(), // 7 days
		"iat":        time.Now().Unix(),
		"token_type": "refresh",
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := refreshTokenObj.SignedString(u.jwtSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

type UserOption func(*userUseCase)

// WithoutRateLimiting disables rate limiting for testing
func WithoutRateLimiting() UserOption {
	return func(uc *userUseCase) {
		uc.disableLimiter = true
	}
}
