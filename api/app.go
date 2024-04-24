package api

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"nfs002/template/v1/internal/db"
	m "nfs002/template/v1/internal/models"
	u "nfs002/template/v1/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
}

type application struct {
	config    config
	validator *validator.Validate
	version   string
	DB        m.DBModel
}

func (app *application) serve() error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		Handler:           app.routes(),
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	log.Info().Int("port", app.config.port).Str("environment", app.config.env).Msg("Starting API Server")

	return srv.ListenAndServe()
}

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterValidation("scope", u.ValidateScope)
	return v
}

// Run API
func Run() {

	// Initialise loggers

	var cfg config

	// DB URL
	cfg.db.dsn = os.Getenv("POSTGRESQL_URL")

	// Port
	cfg.port = u.GetIntEnvOrDefault("API_PORT", 4001)

	// Environment
	cfg.env = u.GetEnvOrDefault("APP_ENV", "dev")

	conn, err := db.OpenDB(cfg.db.dsn)

	if err != nil {
		log.Panic().Str("dsn", cfg.db.dsn).AnErr("error", err).Msg("Failed to open database")
	}

	if cfg.env == "dev" {
		u.InfoLog("Succesfully opened database")
	}

	defer conn.Close()

	app := &application{
		config:    cfg,
		version:   u.API_VERSION,
		DB:        m.DBModel{DB: conn},
		validator: newValidator(),
	}

	err = app.serve()
	if err != nil {
		u.PanicLog("Failed to start API server", err)
	}
}
