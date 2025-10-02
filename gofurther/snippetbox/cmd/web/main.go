package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"

	"github.com/ByChanderZap/snippetbox/internal/models"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql" // New import
)

type application struct {
	logger         *slog.Logger
	snippets       *models.SnippetModel
	templatesCache map[string]*template.Template
	formDecoder    *form.Decoder
}

func main() {
	// this can be setted while running the program like this: go run ./cmd/web -addr=":9999"
	addr := flag.String("addr", ":4000", "Port of where the server will run at")
	dsn := flag.String("dsn", "web:password@tcp(127.0.0.1:3306)/snippetbox?parseTime=true", "MySQL data source name")
	flag.Parse()

	// i might want to read a debug flag to then show logs with debug level
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := openDb(*dsn)
	logger.Info("Connecting to database")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	logger.Info("Database connection stablished")
	defer db.Close()

	// initialize template cache
	tCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	fDecoder := form.NewDecoder()

	app := &application{
		logger:         logger,
		snippets:       &models.SnippetModel{DB: db},
		templatesCache: tCache,
		formDecoder:    fDecoder,
	}

	logger.Info("starting server", "addr", *addr)

	err = http.ListenAndServe(*addr, app.routes())
	if err != nil {
		logger.Error(err.Error())
	}
	os.Exit(1)
}

func openDb(dsn string) (*sql.DB, error) {
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
