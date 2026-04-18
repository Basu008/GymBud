package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Basu008/GymBud/api"
	"github.com/Basu008/GymBud/app"
	"github.com/Basu008/GymBud/server/config"
	"github.com/Basu008/GymBud/server/logger"
	"github.com/Basu008/GymBud/server/middleware"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/urfave/negroni/v3"
)

/*
In this package we will set up our server. Here we will get config, connect to all the DBs, set up our services, start logger, etc
*/

type Server struct {
	httpServer *http.Server
	Router     *mux.Router
	Log        *zerolog.Logger
	Config     *config.Config

	//Storages

	API *api.API
}

func NewServer() *Server {
	c := config.GetConfig()

	server := Server{
		httpServer: &http.Server{},
		Config:     c,
	}

	server.InitLogger()

	r := mux.NewRouter()
	server.Router = r

	appLogger := server.Log.With().Str("type", "app").Logger()
	a := app.NewApp(&app.Options{
		Logger: &appLogger,
		Config: c,
	})

	apiLogger := server.Log.With().Str("type", "api").Logger()
	server.API = api.NewApi(&api.Options{
		MainRouter: r,
		Logger:     &apiLogger,
		Config:     c,
		App:        a,
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
	s.Log.Debug().Msg("Server shut down complete")
}
