package server

import (
	"net/http"

	"github.com/Basu008/GymBud/api"
	"github.com/Basu008/GymBud/server/config"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
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

func NewServer()

func InitLogger()

func StartServer()

func StopServer()
