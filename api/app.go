package api

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"nfs002/template/v1/internal/db"
	m "nfs002/template/v1/internal/models"
	u "nfs002/template/v1/internal/utils"

	"github.com/go-playground/validator/v10"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
	stripe struct {
		secret string
		key    string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
	}
	secretKey      string
	frontend       string
	invoiceService string
}

type application struct {
	config    config
	validator *validator.Validate
	infoLog   *log.Logger
	errorLog  *log.Logger
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

	app.infoLog.Printf("Starting API server in %s mode on port %d", app.config.env, app.config.port)

	return srv.ListenAndServe()
}

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterValidation("scope", u.ValidateScope)
	return v
}

func Run() {

	// Load environment
	infoLog, errorLog := u.LoadEnv()

	var cfg config

	// DB URL
	cfg.db.dsn = os.Getenv("POSTGRESQL_URL")

	flag.StringVar(&cfg.smtp.host, "smtphost", "smtp.mailtrap.io", "smtp host")
	flag.IntVar(&cfg.smtp.port, "smtpport", 587, "smtp port")
	flag.StringVar(&cfg.frontend, "frontend", "http://localhost:4000", "url to front end")
	flag.StringVar(&cfg.invoiceService, "invoice microservice", "http://localhost:5000", "url to invoice microservice")
	flag.Parse()

	// Port
	cfg.port = u.GetIntEnvOrDefault("API_PORT", 4001)

	// Environment
	cfg.env = u.GetEnvOrDefault("API_ENV", "development")

	// stripe
	cfg.stripe.key = os.Getenv("STRIPE_KEY")
	cfg.stripe.secret = os.Getenv("STRIPE_SECRET")

	// smtp
	cfg.smtp.username = os.Getenv("SMTP_USERNAME")
	cfg.smtp.password = os.Getenv("SMTP_PASSWORD")

	// Secret Key (256 bits/32 chars)
	cfg.secretKey = os.Getenv("SECRET_KEY")

	conn, err := db.OpenDB(cfg.db.dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	if cfg.env == "development" {
		infoLog.Printf("Connected to DB @ %s", cfg.db.dsn)
	}
	defer conn.Close()

	app := &application{
		config:    cfg,
		infoLog:   infoLog,
		errorLog:  errorLog,
		version:   version,
		DB:        m.DBModel{DB: conn},
		validator: newValidator(),
	}

	err = app.serve()
	if err != nil {
		app.errorLog.Println(err)
		log.Fatal(err)
	}
}
