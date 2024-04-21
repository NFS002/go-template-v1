package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	// CORS Middleware
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	if app.config.env == "development" {
		mux.Use(middleware.Logger)
	}

	mux.Get("/hello", app.Hello)
	mux.Post("/api/authenticate", app.CreateAuthToken)

	mux.Route("/api/read-a", func(mux chi.Router) {
		mux.Use(app.WithScope([]string{"read:a"}))
		mux.Get("/hello-user", app.HelloUser)
	})

	mux.Route("/api/read-a-write-a", func(mux chi.Router) {
		mux.Use(app.WithScope([]string{"read:a", "write:a"}))
		mux.Get("/hello-user", app.HelloUser)
	})

	// protected routes
	mux.Route("/api/admin", func(mux chi.Router) {
		mux.Use(app.WithScope([]string{"read:a", "write:a", "read:b", "write:b"}))
		mux.Get("/hello-user", app.HelloUser)
		mux.Get("/users", app.GetAllUsers)
		mux.Post("/users", app.CreateUser)
		mux.Get("/users/{id}", app.GetOneUser)
		mux.Put("/users/{id}", app.UpdateUser)
		mux.Delete("/users/{id}", app.DeleteUser)

	})

	return mux
}
