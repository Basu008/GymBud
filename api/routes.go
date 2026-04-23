package api

func (a *API) InitRoutes() {
	a.Router.Root.Handle("/health-check", a.requestHandler(a.healthCheck)).Methods("GET")

	//Auth
	a.Router.APIRoot.Handle("/auth/signup", a.requestHandler(a.signUp)).Methods("POST")
	a.Router.APIRoot.Handle("/auth/login", a.requestHandler(a.login)).Methods("POST")
	a.Router.APIRoot.Handle("/auth/logout", a.requestAuthHandler(a.logout)).Methods("POST")

	//Users
	a.Router.APIRoot.Handle("/users/me", a.requestAuthHandler(a.getCurrentUser)).Methods("GET")
	a.Router.APIRoot.Handle("/users/{id}", a.requestAuthHandler(a.getUserByID)).Methods("GET")
	a.Router.APIRoot.Handle("/users/{id}/follow", a.requestAuthHandler(a.followUser)).Methods("POST")
	a.Router.APIRoot.Handle("/users/{id}/follow", a.requestAuthHandler(a.unfollowUser)).Methods("DELETE")
	a.Router.APIRoot.Handle("/users/me", a.requestAuthHandler(a.updateUser)).Methods("PATCH")
	a.Router.APIRoot.Handle("/users/me/privacy", a.requestAuthHandler(a.updatePrivacy)).Methods("PATCH")
	a.Router.APIRoot.Handle("/users/me/active", a.requestAuthHandler(a.updateActive)).Methods("PATCH")
	a.Router.APIRoot.Handle("/users/me/body-metrics", a.requestAuthHandler(a.createBodyMetrics)).Methods("POST")
	a.Router.APIRoot.Handle("/users/me/body-metrics/{id}", a.requestAuthHandler(a.deleteBodyMetrics)).Methods("DELETE")

	//Exercises
	a.Router.APIRoot.Handle("/exercises", a.requestAuthHandler(a.createExercise)).Methods("POST")
	a.Router.APIRoot.Handle("/exercises/{id}", a.requestAuthHandler(a.updateExercise)).Methods("PATCH")
	a.Router.APIRoot.Handle("/exercises/{id}", a.requestAuthHandler(a.deleteExercise)).Methods("DELETE")
}
