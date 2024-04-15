package gapi

import (
	"context"

	"github.com/Brysen/simplebank/db/sqlc"
	"github.com/Brysen/simplebank/pb"
	"github.com/Brysen/simplebank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) VerifyEmail(ctx context.Context, request *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	violations := validateVerifyEmailRequest(request)
	if violations != nil {
		return nil, InvalidArgumentError(violations)
	}

	txResult, err := server.store.VerifyEmailTx(ctx, sqlc.VerifyEmailTxParams{
		EmailId:    request.GetEmailId(),
		SecretCode: request.GetSecretCode(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify email")
	}

	rsp := &pb.VerifyEmailResponse{
		IsVerified: txResult.User.IsEmailVerified,
	}
	return rsp, nil
}

func validateVerifyEmailRequest(request *pb.VerifyEmailRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateEmailId(request.GetEmailId()); err != nil {
		violations = append(violations, fieldViolation("email_id", err))
	}

	if err := val.ValidateSecretCode(request.GetSecretCode()); err != nil {
		violations = append(violations, fieldViolation("secret_code", err))
	}

	return violations
}
