package handler

import (
	"context"

	"github.com/pt-xyz-multifinance/internal/model"
	"github.com/pt-xyz-multifinance/internal/usecase"
	pb "github.com/pt-xyz-multifinance/proto/gen/go/xyz/multifinance/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LoanHandler struct {
	pb.UnimplementedLoanServiceServer
	loanUseCase usecase.LoanUseCase
	log         *zap.Logger
}

func NewLoanHandler(loanUseCase usecase.LoanUseCase, log *zap.Logger) *LoanHandler {
	return &LoanHandler{
		loanUseCase: loanUseCase,
		log:         log,
	}
}

func (h *LoanHandler) ApplyLoan(ctx context.Context, req *pb.LoanApplicationRequest) (*pb.LoanApplication, error) {
	loan, err := h.loanUseCase.ApplyLoan(ctx, req.UserId, req.Amount, int(req.TenureMonths), req.Purpose)
	if err != nil {
		h.log.Error("Failed to apply loan", zap.Error(err))
		return nil, err
	}

	return convertLoanToProto(loan), nil
}

func (h *LoanHandler) GetLoanStatus(ctx context.Context, req *pb.GetLoanStatusRequest) (*pb.LoanApplication, error) {
	loan, err := h.loanUseCase.GetLoanStatus(ctx, req.LoanId)
	if err != nil {
		h.log.Error("Failed to get loan status", zap.Error(err))
		return nil, err
	}

	return convertLoanToProto(loan), nil
}

func (h *LoanHandler) GetLoanHistory(ctx context.Context, req *pb.GetLoanHistoryRequest) (*pb.GetLoanHistoryResponse, error) {
	loans, total, err := h.loanUseCase.GetLoanHistory(ctx, req.UserId, int(req.Page), int(req.PageSize))
	if err != nil {
		h.log.Error("Failed to get loan history", zap.Error(err))
		return nil, err
	}

	response := &pb.GetLoanHistoryResponse{
		Loans:    make([]*pb.LoanApplication, 0, len(loans)),
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	for _, loan := range loans {
		response.Loans = append(response.Loans, convertLoanToProto(&loan))
	}

	return response, nil
}

func (h *LoanHandler) SubmitLoanDocuments(ctx context.Context, req *pb.SubmitLoanDocumentsRequest) (*pb.LoanApplication, error) {
	docs := make([]model.Document, 0, len(req.Documents))
	for _, doc := range req.Documents {
		docs = append(docs, model.Document{
			Type: model.DocumentType(doc.Type),
			Name: doc.Name,
			URL:  doc.Url,
		})
	}

	if err := h.loanUseCase.SubmitLoanDocuments(ctx, req.LoanId, docs); err != nil {
		h.log.Error("Failed to submit loan documents", zap.Error(err))
		return nil, err
	}

	loan, err := h.loanUseCase.GetLoanStatus(ctx, req.LoanId)
	if err != nil {
		h.log.Error("Failed to get updated loan status", zap.Error(err))
		return nil, err
	}

	return convertLoanToProto(loan), nil
}

// Helper function to convert model.Loan to proto LoanApplication
func convertLoanToProto(loan *model.Loan) *pb.LoanApplication {
	if loan == nil {
		return nil
	}

	result := &pb.LoanApplication{
		Id:             loan.ID,
		UserId:         loan.UserID,
		Amount:         loan.Amount,
		TenureMonths:   int32(loan.TenureMonths),
		Purpose:        loan.Purpose,
		Status:         string(loan.Status),
		MonthlyPayment: loan.MonthlyPayment,
		InterestRate:   loan.InterestRate,
		CreatedAt:      timestamppb.New(loan.CreatedAt),
		UpdatedAt:      timestamppb.New(loan.UpdatedAt),
	}

	if loan.DisbursedAmount > 0 {
		result.DisbursedAmount = loan.DisbursedAmount
	}
	if loan.DisbursedAt != nil {
		result.DisbursedAt = timestamppb.New(*loan.DisbursedAt)
	}

	result.Documents = make([]*pb.Document, 0, len(loan.Documents))
	for _, doc := range loan.Documents {
		pdoc := &pb.Document{
			Id:     doc.ID,
			Type:   string(doc.Type),
			Name:   doc.Name,
			Status: string(doc.Status),
			Url:    doc.URL,
		}
		if doc.UploadedAt != nil {
			pdoc.UploadedAt = timestamppb.New(*doc.UploadedAt)
		}
		result.Documents = append(result.Documents, pdoc)
	}

	return result
}
