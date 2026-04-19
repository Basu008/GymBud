package api

import (
	"net/http"
	"sync"

	"github.com/Basu008/GymBud/app"
	"github.com/Basu008/GymBud/server/config"
	"github.com/Basu008/GymBud/server/handler"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type API struct {
	Router     *Router
	MainRouter *mux.Router
	Logger     *zerolog.Logger
	Config     *config.Config
	TokenAuth  handler.TokenAuth
	Mutex      *sync.Mutex
	App        *app.App
}

type Options struct {
	MainRouter *mux.Router
	Logger     *zerolog.Logger
	Config     *config.Config
	TokenAuth  handler.TokenAuth
	App        *app.App
}

type Router struct {
	Root    *mux.Router
	APIRoot *mux.Router
}

func NewApi(opts *Options) *API {
	api := API{
		Router:     &Router{},
		MainRouter: opts.MainRouter,
		Logger:     opts.Logger,
		TokenAuth:  opts.TokenAuth,
		Config:     opts.Config,
		Mutex:      &sync.Mutex{},
		App:        opts.App,
	}
	api.setUpRoutes()
	return &api
}

func (a *API) setUpRoutes() {
	a.Router.Root = a.MainRouter
	a.Router.APIRoot = a.MainRouter.PathPrefix("/v1/api").Subrouter()
	a.InitRoutes()
}

func (a *API) requestHandler(h func(*handler.RequestCtx, http.ResponseWriter, *http.Request)) http.Handler {
	return a.wrapHandler(h, false)
}

func (a *API) requestAuthHandler(h func(*handler.RequestCtx, http.ResponseWriter, *http.Request)) http.Handler {
	return a.wrapHandler(h, true)
}

func (a *API) wrapHandler(h func(*handler.RequestCtx, http.ResponseWriter, *http.Request), isLoggedIn bool) http.Handler {
	var authFunc handler.TokenAuth
	if a.TokenAuth != nil {
		authFunc = a.TokenAuth
	}
	return &handler.Request{
		HandlerFunc: h,
		AuthFunc:    authFunc,
		IsLoggedIn:  isLoggedIn,
	}
}
