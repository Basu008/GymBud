package api

func (a *API) InitRoutes() {
	a.Router.Root.Handle("/health-check", a.requestHandler(a.healthCheck)).Methods("GET")
}
