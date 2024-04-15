package gapi

import (
	"context"
	"time"

	"github.com/Brysen/simplebank/db/sqlc"
	"github.com/Brysen/simplebank/pb"
	"github.com/Brysen/simplebank/util"
	"github.com/Brysen/simplebank/val"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) UpdateUser(ctx context.Context, request *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	//TODO: add authorization
	authPayload, err := server.authorizeUser(ctx, []string{util.BankerRole, util.DepositorRole})
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateUpdateUserRequest(request)
	if violations != nil {
		return nil, InvalidArgumentError(violations)
	}

	if authPayload.Role != util.BankerRole && authPayload.Username != request.GetUsername() {
		return nil, status.Errorf(codes.PermissionDenied, "cannot update other's info")
	}

	arg := sqlc.UpdateUserParams{
		Username: request.GetUsername(),
		FullName: pgtype.Text{
			String: request.GetFullName(),
			Valid:  request.FullName != nil,
		},
		Email: pgtype.Text{
			String: request.GetEmail(),
			Valid:  request.Email != nil,
		},
	}

	if request.Password != nil {
		hashedPassword, err := util.HashPassword(request.GetPassword())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to hash password %s", err)
		}
		arg.HashedPassword = pgtype.Text{
			String: hashedPassword,
			Valid:  true,
		}

		arg.PasswordChangedAt = pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		}
	}

	user, err := server.store.UpdateUser(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update user %s", err)
	}

	rsp := &pb.UpdateUserResponse{
		User: convertUser(user),
	}
	return rsp, nil
}

func validateUpdateUserRequest(request *pb.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(request.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if request.Password != nil {
		if err := val.ValidatePassword(request.GetPassword()); err != nil {
			violations = append(violations, fieldViolation("full_name", err))
		}
	}

	if request.FullName != nil {
		if err := val.ValidateFullName(request.GetFullName()); err != nil {
			violations = append(violations, fieldViolation("full_name", err))
		}
	}

	if request.Email != nil {
		if err := val.ValidateEmail(request.GetEmail()); err != nil {
			violations = append(violations, fieldViolation("email", err))
		}
	}

	return violations
}
