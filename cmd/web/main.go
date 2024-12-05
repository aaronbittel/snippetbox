package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/aaronbittel/snippetbox/internal/models"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	_ "github.com/go-sql-driver/mysql"
)

type config struct {
	addr      string
	staticDir string
	dsn       string
}

type application struct {
	users          *models.UserModel
	snippets       *models.SnippetModel
	logger         *slog.Logger
	templateCache  map[string]*template.Template
	cfg            config
	sessionManager *scs.SessionManager
}

func main() {
	var cfg config

	flag.StringVar(&cfg.addr, "addr", ":4000", "HTTP network address")
	flag.StringVar(&cfg.staticDir, "static-dir", "./ui/static/", "Path to static assets")
	flag.StringVar(&cfg.dsn, "dsn",
		"web:snippet@/snippetbox?parseTime=true",
		"MySQL data source name")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	db, err := openDB(cfg.dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	app := &application{
		logger:         logger,
		cfg:            cfg,
		snippets:       &models.SnippetModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		sessionManager: sessionManager,
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := http.Server{
		Addr:         cfg.addr,
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	logger.Info("starting server", "addr", cfg.addr)

	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	logger.Error(err.Error())
	os.Exit(1)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
