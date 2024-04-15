package gapi

import (
	"testing"
	"time"

	"github.com/Brysen/simplebank/db/sqlc"
	"github.com/Brysen/simplebank/util"
	"github.com/Brysen/simplebank/worker"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, store sqlc.Store, taskDistributor worker.TaskDistributor) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}
	server, err := NewServer(config, store, taskDistributor)
	require.NoError(t, err)

	return server

}
