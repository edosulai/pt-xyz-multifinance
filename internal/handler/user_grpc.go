package handler

import (
	"context"
	"errors"

	"github.com/edosulai/pt-xyz-multifinance/internal/model"
	"github.com/edosulai/pt-xyz-multifinance/internal/usecase"
	pb "github.com/edosulai/pt-xyz-multifinance/proto/gen/go/xyz/multifinance/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	userUseCase usecase.UserUseCase
	logger      *zap.Logger
}

func NewUserHandler(userUseCase usecase.UserUseCase, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
		logger:      logger,
	}
}

func (h *UserHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user := &model.User{
		Username:      req.Username,
		Email:         req.Email,
		Password:      req.Password,
		FullName:      req.FullName,
		PhoneNumber:   req.PhoneNumber,
		Address:       req.Address,
		KTPNumber:     req.KtpNumber,
		MonthlyIncome: req.MonthlyIncome,
		Status:        "active",
	}

	err := h.userUseCase.Register(ctx, user)
	if err != nil {
		code := codes.Internal
		msg := "failed to register user"

		switch {
		case errors.As(err, &usecase.ValidationError{}):
			code = codes.InvalidArgument
			msg = err.Error()
		case errors.As(err, &usecase.ConflictError{}):
			code = codes.AlreadyExists
			msg = err.Error()
		default:
			h.logger.Error("Failed to register user",
				zap.Error(err),
				zap.String("username", user.Username),
				zap.String("email", user.Email))
		}

		return nil, status.Error(code, msg)
	}

	return &pb.RegisterResponse{
		User: convertUserToUserInfo(user),
	}, nil
}

func (h *UserHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Validate captcha
	if !h.userUseCase.ValidateCaptcha(req.GetCaptchaId(), req.GetCaptchaSolution()) {
		return nil, status.Error(codes.InvalidArgument, "invalid captcha")
	}

	token, refreshToken, user, err := h.userUseCase.Login(ctx, req.Username, req.Password)
	if err != nil {
		if err == usecase.ErrInvalidCredentials {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		h.logger.Error("Failed to login user", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to login user")
	}

	return &pb.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User: &pb.UserInfo{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FullName:  user.FullName,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}, nil
}

func (h *UserHandler) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.GetProfileResponse, error) {
	user, err := h.userUseCase.GetProfile(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to get user profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get user profile")
	}

	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &pb.GetProfileResponse{
		User: convertUserToUserInfo(user),
	}, nil
}

func (h *UserHandler) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UserInfo, error) {
	user, err := h.userUseCase.GetProfile(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to get user profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get user profile")
	}

	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Update user fields
	user.PhoneNumber = req.PhoneNumber
	user.Address = req.Address
	user.KTPNumber = req.KtpNumber
	user.FullName = req.FullName
	user.MonthlyIncome = req.MonthlyIncome

	err = h.userUseCase.UpdateProfile(ctx, user)
	if err != nil {
		h.logger.Error("Failed to update user profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update user profile")
	}

	return convertUserToUserInfo(user), nil
}

// Helper function to convert model.User to pb.UserInfo
func convertUserToUserInfo(user *model.User) *pb.UserInfo {
	if user == nil {
		return nil
	}

	return &pb.UserInfo{
		Id:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		FullName:      user.FullName,
		PhoneNumber:   user.PhoneNumber,
		Address:       user.Address,
		KtpNumber:     user.KTPNumber,
		Status:        user.Status,
		MonthlyIncome: user.MonthlyIncome,
		CreatedAt:     timestamppb.New(user.CreatedAt),
		UpdatedAt:     timestamppb.New(user.UpdatedAt),
	}
}
