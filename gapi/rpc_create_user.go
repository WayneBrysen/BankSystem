package gapi

import (
	"context"
	"time"

	"github.com/Brysen/simplebank/db/sqlc"
	"github.com/Brysen/simplebank/pb"
	"github.com/Brysen/simplebank/util"
	"github.com/Brysen/simplebank/val"
	"github.com/Brysen/simplebank/worker"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateUser(ctx context.Context, request *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	
	violations := validateCreateUserRequest(request)
	if violations != nil {
		return nil, InvalidArgumentError(violations)
	}

	hashedPassword, err := util.HashPassword(request.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password %s", err)
	}

	arg := sqlc.CreateUserTxParams{
		CreateUserParams: sqlc.CreateUserParams{
			Username:       request.GetUsername(),
			HashedPassword: hashedPassword,
			FullName:       request.GetFullName(),
			Email:          request.GetEmail(),
		},
		AfterCreate: func(user sqlc.User) error {
			taskPayload := &worker.PayloadSendVerifyEmail{
				Username: user.Username,
			}

			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}

			return server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...)
		},
	}

	txResult, err := server.store.CreateUserTx(ctx, arg)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.Code {
			case "unique_violation":
				return nil, status.Errorf(codes.AlreadyExists, "username already exists %s", err)
			}
		}
		return nil, status.Errorf(codes.Internal, "failed to create user %s", err)
	}

	// TODO: use db transaction

	rsp := &pb.CreateUserResponse{
		User: convertUser(txResult.User),
	}
	return rsp, nil
}

func validateCreateUserRequest(request *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(request.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if err := val.ValidatePassword(request.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("full_name", err))
	}

	if err := val.ValidateFullName(request.GetFullName()); err != nil {
		violations = append(violations, fieldViolation("full_name", err))
	}

	if err := val.ValidateEmail(request.GetEmail()); err != nil {
		violations = append(violations, fieldViolation("email", err))
	}

	return violations
}
