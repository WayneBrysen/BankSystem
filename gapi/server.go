package gapi

import (
	"fmt"

	"github.com/Brysen/simplebank/db/sqlc"
	"github.com/Brysen/simplebank/pb"
	"github.com/Brysen/simplebank/token"
	"github.com/Brysen/simplebank/util"
	"github.com/Brysen/simplebank/worker"
)

// Server serves gRPC requests for our banking service.
type Server struct {
	pb.UnimplementedSimplebankServer
	config          util.Config
	store           sqlc.Store
	tokenMaker      token.Maker
	taskDistributor worker.TaskDistributor
}

func NewServer(config util.Config, store sqlc.Store, taskDistributor worker.TaskDistributor) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:          config,
		store:           store,
		tokenMaker:      tokenMaker,
		taskDistributor: taskDistributor,
	}

	return server, nil
}
