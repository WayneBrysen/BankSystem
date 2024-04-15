package gapi

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	mockdb "github.com/Brysen/simplebank/db/mock"
	"github.com/Brysen/simplebank/db/sqlc"
	"github.com/Brysen/simplebank/pb"
	"github.com/Brysen/simplebank/util"
	mockwk "github.com/Brysen/simplebank/worker/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type eqCreateUserTxParamsMatcher struct {
	arg      sqlc.CreateUserTxParams
	password string
}

func (expected eqCreateUserTxParamsMatcher) Matches(x interface{}) bool {
	actualArg, ok := x.(sqlc.CreateUserTxParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(expected.password, actualArg.HashedPassword)
	if err != nil {
		return false
	}

	expected.arg.HashedPassword = actualArg.HashedPassword
	return reflect.DeepEqual(expected.arg.CreateUserParams, actualArg.CreateUserParams)

}

func (e eqCreateUserTxParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserTxParams(arg sqlc.CreateUserTxParams, password string) gomock.Matcher {
	return eqCreateUserTxParamsMatcher{arg, password}
}

func randomUser(t *testing.T) (user sqlc.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = sqlc.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	return
}

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		req           *pb.CreateUserRequest
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, res *pb.CreateUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := sqlc.CreateUserTxParams{
					CreateUserParams: sqlc.CreateUserParams{
						Username: user.Username,
						FullName: user.FullName,
						Email:    user.Email,
					},
				}
				store.EXPECT().CreateUserTx(gomock.Any(), EqCreateUserTxParams(arg, password)).Times(1).Return(sqlc.CreateUserTxResult{User: user}, nil)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				createdUser := res.GetUser()
				require.Equal(t, user.Username, createdUser.Username)
				require.Equal(t, user.FullName, createdUser.FullName)
				require.Equal(t, user.Email, createdUser.Email)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			taskDistributor := mockwk.NewMockTaskDistributor(ctrl)

			server := newTestServer(t, store, taskDistributor)
			res, err := server.CreateUser(context.Background(), tc.req)
			tc.checkResponse(t, res, err)

		})

	}

}
