package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/Basu008/GymBud/api"
	"github.com/Basu008/GymBud/app"
	"github.com/Basu008/GymBud/server/auth"
	"github.com/Basu008/GymBud/server/config"
	"github.com/Basu008/GymBud/server/database"
	"github.com/Basu008/GymBud/server/logger"
	"github.com/Basu008/GymBud/server/middleware"
	"github.com/Basu008/GymBud/server/validator"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/urfave/negroni/v3"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

/*
In this package we will set up our server. Here we will get config, connect to all the DBs, set up our services, start logger, etc
*/

type Server struct {
	httpServer *http.Server
	Router     *mux.Router
	Log        *zerolog.Logger
	Config     *config.Config

	Mongo    *mongo.Client
	Postgres *pgxpool.Pool
	Redis    *redis.Client
	Firebase *firebase.App

	API *api.API
}

func NewServer() *Server {
	c := config.GetConfig()

	server := Server{
		httpServer: &http.Server{},
		Config:     c,
	}

	server.InitLogger()

	server.Mongo = database.NewMongoClient(c.MongoDatabaseConfig)
	server.Postgres = database.NewPostgresPool(c.PostgresDatabaseConfig)
	server.Redis = database.NewRedisClient(c.RedisConfig)
	server.Firebase = database.NewFireBaseApp(&c.FirebaseConfig)

	r := mux.NewRouter()
	server.Router = r
	authService := auth.NewAuthService(&auth.Options{Config: c, Redis: server.Redis})

	appLogger := server.Log.With().Str("type", "app").Logger()
	a := app.NewApp(&app.Options{
		Logger:   &appLogger,
		Config:   c,
		Mongo:    server.Mongo.Database(c.MongoDatabaseConfig.Database),
		Postgres: server.Postgres,
		Redis:    server.Redis,
		Firebase: server.Firebase,
	})
	a.AuthService = authService
	if err := app.InitService(a); err != nil {
		server.Log.Fatal().Err(err).Msg("failed to initialize services")
	}

	apiLogger := server.Log.With().Str("type", "api").Logger()
	server.API = api.NewApi(&api.Options{
		MainRouter: r,
		Logger:     &apiLogger,
		Config:     c,
		App:        a,
		Validator:  validator.NewValidator(),
		TokenAuth:  authService,
	})
	return &server
}

func (s *Server) InitLogger() {
	l := logger.NewLogger()
	s.Log = &l
}

func (s *Server) StartServer() {
	n := negroni.New()

	serverConfig := s.Config.ServerConfig
	c := cors.New(cors.Options{
		AllowedOrigins:   serverConfig.CORSAllowedOrigins,
		AllowedMethods:   serverConfig.CORSAllowedMethods,
		AllowedHeaders:   serverConfig.CORSAllowedHeaders,
		AllowCredentials: true,
	})

	recovery := negroni.NewRecovery()
	n.Use(recovery)
	n.Use(c)
	requestLogger := s.Log.With().Str("type", "request").Logger()
	n.Use(middleware.NewRequestLoggerWithLogger(&requestLogger))
	n.UseHandler(s.Router)

	address := serverConfig.ListenAddress
	if address == "" {
		address = ":" + serverConfig.Port
	}
	readTimeout := serverConfig.ReadTimeout
	if readTimeout == 0 {
		readTimeout = 150 * time.Second
	}
	writeTimeout := serverConfig.WriteTimeout
	if writeTimeout == 0 {
		writeTimeout = 150 * time.Second
	}

	s.httpServer = &http.Server{
		Handler:      n,
		Addr:         fmt.Sprintf("%s:%s", address, serverConfig.Port),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	s.Log.Info().Msgf("Starting server at %s", s.httpServer.Addr)
	go func() {
		err := s.httpServer.ListenAndServe()
		if err != nil {
			s.Log.Error().Err(err).Msg("")
			return
		}
	}()
}

func (s *Server) StopServer() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	s.Log.Debug().Msg("Shutting down server")
	s.httpServer.Shutdown(ctx)
	if s.Mongo != nil {
		if err := s.Mongo.Disconnect(ctx); err != nil {
			s.Log.Error().Err(err).Msg("failed to disconnect mongo client")
		}
	}
	if s.Postgres != nil {
		s.Postgres.Close()
	}
	if s.Redis != nil {
		if err := s.Redis.Close(); err != nil {
			s.Log.Error().Err(err).Msg("failed to close redis client")
		}
	}
	s.Log.Debug().Msg("Server shut down complete")
}
