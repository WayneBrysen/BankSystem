package sqlc

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/Brysen/simplebank/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testQueries *Queries
var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	testDB, err = pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	defer testDB.Close()

	testQueries = New(testDB)

	os.Exit(m.Run())
}
