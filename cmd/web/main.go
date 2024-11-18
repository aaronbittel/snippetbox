package main

import (
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"

	"github.com/aaronbittel/snippetbox/internal/models"
	_ "github.com/go-sql-driver/mysql"
)

type config struct {
	addr      string
	staticDir string
	dsn       string
}

type application struct {
	snippets *models.SnippetModel
	logger   *slog.Logger
	cfg      config
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	var cfg config

	flag.StringVar(&cfg.addr, "addr", ":4000", "HTTP network address")
	flag.StringVar(&cfg.staticDir, "static-dir", "./ui/static/", "Path to static assets")
	flag.StringVar(&cfg.dsn,
		"dsn",
		"web:snippet@/snippetbox?parseTime=true",
		"MySQL data source name")

	flag.Parse()

	db, err := openDB(cfg.dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	app := &application{
		logger:   logger,
		cfg:      cfg,
		snippets: &models.SnippetModel{DB: db},
	}

	logger.Info("starting server", "addr", cfg.addr)

	mux := app.routes()

	err = http.ListenAndServe(cfg.addr, mux)
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
