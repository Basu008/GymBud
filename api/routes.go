package api

func (a *API) InitRoutes() {
	a.Router.Root.Handle("/health-check", a.requestHandler(a.healthCheck)).Methods("GET")

	//Users
	a.Router.APIRoot.Handle("/auth/signup", a.requestHandler(a.signUp)).Methods("POST")
}
