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
	a.Router.APIRoot.Handle("/users/me", a.requestAuthHandler(a.updateUser)).Methods("PATCH")
	a.Router.APIRoot.Handle("/users/me/privacy", a.requestAuthHandler(a.updatePrivacy)).Methods("PATCH")
	a.Router.APIRoot.Handle("/users/me/active", a.requestAuthHandler(a.updateActive)).Methods("PATCH")
	a.Router.APIRoot.Handle("/users/me/body-metrics", a.requestAuthHandler(a.createBodyMetrics)).Methods("POST")
	a.Router.APIRoot.Handle("/users/me/body-metrics/{id}", a.requestAuthHandler(a.deleteBodyMetrics)).Methods("DELETE")
	a.Router.APIRoot.Handle("/users/me/workouts", a.requestAuthHandler(a.listCurrentUserWorkouts)).Methods("GET")
	a.Router.APIRoot.Handle("/users/me/following/workouts", a.requestAuthHandler(a.listFollowingWorkouts)).Methods("GET")
	a.Router.APIRoot.Handle("/users/me/workouts/analytics", a.requestAuthHandler(a.getCurrentUserWorkoutAnalytics)).Methods("GET")
	a.Router.APIRoot.Handle("/users/{id}/workouts", a.requestAuthHandler(a.listUserWorkouts)).Methods("GET")
	a.Router.APIRoot.Handle("/users/{id}/workouts/analytics", a.requestAuthHandler(a.getUserWorkoutAnalytics)).Methods("GET")

	//Social
	a.Router.APIRoot.Handle("/users/{id}/follow", a.requestAuthHandler(a.followUser)).Methods("POST")
	a.Router.APIRoot.Handle("/users/{id}/follow", a.requestAuthHandler(a.unfollowUser)).Methods("DELETE")
	a.Router.APIRoot.Handle("/users/{id}/follow/accept", a.requestAuthHandler(a.acceptFollowRequest)).Methods("POST")
	a.Router.APIRoot.Handle("/users/{id}/follow/reject", a.requestAuthHandler(a.rejectFollowRequest)).Methods("DELETE")

	//Exercises
	a.Router.APIRoot.Handle("/exercises", a.requestAuthHandler(a.listExercises)).Methods("GET")
	a.Router.APIRoot.Handle("/exercises", a.requestAuthHandler(a.createExercise)).Methods("POST")
	a.Router.Root.Handle("/exercises/bulk", a.requestHandler(a.createExercises)).Methods("POST")
	a.Router.APIRoot.Handle("/exercises/categories", a.requestHandler(a.listExerciseCategories)).Methods("GET")
	a.Router.APIRoot.Handle("/exercises/muscles", a.requestHandler(a.listExerciseMuscles)).Methods("GET")
	a.Router.APIRoot.Handle("/exercises/equipments", a.requestHandler(a.listExerciseEquipments)).Methods("GET")
	a.Router.APIRoot.Handle("/exercises/difficulty", a.requestHandler(a.listExerciseDifficulty)).Methods("GET")
	a.Router.APIRoot.Handle("/exercises/{id}", a.requestAuthHandler(a.getExerciseByID)).Methods("GET")
	a.Router.APIRoot.Handle("/exercises/{id}", a.requestAuthHandler(a.updateExercise)).Methods("PATCH")
	a.Router.APIRoot.Handle("/exercises/{id}", a.requestAuthHandler(a.deleteExercise)).Methods("DELETE")

	//Routines
	a.Router.APIRoot.Handle("/routines", a.requestAuthHandler(a.listRoutines)).Methods("GET")
	a.Router.APIRoot.Handle("/routines", a.requestAuthHandler(a.createRoutine)).Methods("POST")
	a.Router.APIRoot.Handle("/routines/{id}/copy", a.requestAuthHandler(a.copyRoutine)).Methods("POST")
	a.Router.APIRoot.Handle("/routines/{id}", a.requestAuthHandler(a.getRoutineByID)).Methods("GET")
	a.Router.APIRoot.Handle("/routines/{id}/workouts/latest", a.requestAuthHandler(a.getLatestWorkoutByRoutineID)).Methods("GET")
	a.Router.APIRoot.Handle("/routines/{id}", a.requestAuthHandler(a.updateRoutine)).Methods("PATCH")
	a.Router.APIRoot.Handle("/routines/{id}", a.requestAuthHandler(a.deleteRoutine)).Methods("DELETE")

	//Workouts
	a.Router.APIRoot.Handle("/workouts", a.requestAuthHandler(a.createWorkout)).Methods("POST")
	a.Router.APIRoot.Handle("/workouts/{id}", a.requestAuthHandler(a.getWorkoutByID)).Methods("GET")
	a.Router.APIRoot.Handle("/workouts/{id}", a.requestAuthHandler(a.deleteWorkout)).Methods("DELETE")
	a.Router.APIRoot.Handle("/workouts/{id}/like", a.requestAuthHandler(a.likeWorkout)).Methods("POST")
	a.Router.APIRoot.Handle("/workouts/{id}/like", a.requestAuthHandler(a.unlikeWorkout)).Methods("DELETE")

	//Media
	a.Router.APIRoot.Handle("/media/images", a.requestAuthHandler(a.uploadImage)).Methods("POST")
}
