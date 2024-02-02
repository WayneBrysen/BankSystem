package gapi

import (
	"github.com/Brysen/simplebank/db/sqlc"
	"github.com/Brysen/simplebank/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertUser(user sqlc.User) *pb.User {
	return &pb.User{
		Username:         user.Username,
		FullName:         user.FullName,
		Email:            user.Email,
		PasswordChangeAt: timestamppb.New(user.PasswordChangedAt.Time),
		CreatedAt:        timestamppb.New(user.CreatedAt.Time),
	}
}
