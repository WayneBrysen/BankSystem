package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/Brysen/simplebank/api"
	"github.com/Brysen/simplebank/db/sqlc"
	_ "github.com/Brysen/simplebank/doc/statik"
	"github.com/Brysen/simplebank/gapi"
	"github.com/Brysen/simplebank/mail"
	"github.com/Brysen/simplebank/pb"
	"github.com/Brysen/simplebank/util"
	"github.com/Brysen/simplebank/worker"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Msg("cannot load config:")
	}

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()

	conn, err := pgxpool.New(ctx, config.DBSource)
	if err != nil {
		log.Fatal().Msg("cannot connect to db:")
	}

	// run db migration
	runDBMigration(config.MigrationURL, config.DBSource)

	store := sqlc.NewStore(conn)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	waitGroup, ctx := errgroup.WithContext(ctx)

	runTaskProcessor(ctx, waitGroup, config, redisOpt, store)
	runGatewayServer(ctx, waitGroup, config, store, taskDistributor)
	runGrpcServer(ctx, waitGroup, config, store, taskDistributor)

	err = waitGroup.Wait()
	if err != nil {
		log.Fatal().Err(err).Msg("err from wait group")
	}
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Msg("cannot create new migrate instance")
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Msg("failed to run migrate up:")
	}

	log.Info().Msg("db migrated successfully")
}

func runTaskProcessor(ctx context.Context, waitGroup *errgroup.Group, config util.Config, redisOpt asynq.RedisClientOpt, store sqlc.Store) {
	mailer := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, mailer)
	
	log.Info().Msg("start task processor")
	err := taskProcessor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start task processor")
	}

	waitGroup.Go(func () error {
		<-ctx.Done()
		log.Info().Msg("shutdown task processor")

		taskProcessor.Shutdown()
		log.Info().Msg("shutdown task processor")
		return nil
	})
}

func runGrpcServer(ctx context.Context, waitGroup *errgroup.Group, config util.Config, store sqlc.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Msg("cannot create server:")
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterSimplebankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal().Msg("cannot create listener")
	}

	waitGroup.Go(func() error {
		log.Printf("start gRPC server at %s", listener.Addr().String())
		err = grpcServer.Serve(listener)
		if err != nil {
			log.Error().Msg("cannot start gRPC server")
			return err
		}

		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("shutdown gRPC server")
		grpcServer.GracefulStop()
		log.Info().Msg("gRPC server now is stopped")

		return nil
	})

}

func runGatewayServer(ctx context.Context, waitGroup *errgroup.Group, config util.Config, store sqlc.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Msg("cannot create server:")
	}

	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOption)

	err = pb.RegisterSimplebankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Msg("cannot register handler server")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal().Msg("cannot create statik fs")
	}

	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swaggerHandler)

	httpServer := &http.Server {
		Handler: gapi.HttpLogger(mux),
		Addr: config.HTTPServerAddress,
	}

	waitGroup.Go(func() error {
		log.Printf("start HTTP gateway server at %s", httpServer.Addr)
		err = httpServer.ListenAndServe()
		if err != nil {
			log.Error().Err(err).Msg("cannot start HTTP gateway server")
			return err
		}
		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("shutdown HTTP gateway server")

		err := httpServer.Shutdown(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("failed to shutdown HTTP gateway")
			return err
		}

		log.Info().Msg("Http gateway server is stopped")
		return nil
	})



}

func runGinServer(config util.Config, store sqlc.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Msg("cannot create server:")
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Msg("cannot start server:")
	}
}
