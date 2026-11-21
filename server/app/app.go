package app

import (
	"context"
	"flag"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/nhht77/earth-rest-api/pkg/logging"
	"github.com/nhht77/earth-rest-api/server/config"
	"github.com/nhht77/earth-rest-api/server/database"
	"github.com/nhht77/earth-rest-api/server/handlers"
)

type App struct {
	Log      logging.Logger
	DB       *database.Database
	Conf     *config.Config
	Handlers *handlers.Handlers
}

// New builds the App skeleton from the one thing that must exist first:
// the logger. Dependencies are created but NOT started — that's Setup.
func New(log logging.Logger) *App {
	return &App{
		Log:  log,
		Conf: &config.Config{},
		DB:   database.NewDatabase(log),
	}
}

func (a *App) Setup(ctx context.Context) error {
	if err := a.Initialize(os.Args[1:]); err != nil {
		return err
	}
	a.Conf.ReadDefault(*a.Conf, a.Log)
	if err := a.DB.Initialize(a.Conf); err != nil {
		return err
	}
	if a.Conf.Framework.IsTestBuild {
		if err := a.DB.ClearTable(); err != nil {
			return err
		}
	}
	a.Handlers = handlers.NewHandlers(a.DB, a.Log)
	return nil
}

// Run serves until ListenAndServe returns.
func (a *App) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:    ":" + a.Conf.Framework.ServerPort,
		Handler: handlers.RunHTTP(a.Handlers),
	}

	errCh := make(chan error, 1)
	go func() {
		a.Log.Infof("Starting server at port %s", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		a.Log.Infof("Shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
	return nil
}

// Close releases resources — called via defer in main, even on error paths.
func (a *App) Close() {
	a.DB.Close()
}

func (a *App) Initialize(args []string) error {
	fs := flag.NewFlagSet("earth-rest-api", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // keep usage/parse errors out of test output

	dataSource := &config.DataSource{}
	fs.StringVar(&dataSource.DatabaseHost, "database-host", "localhost", "Database host")
	fs.StringVar(&dataSource.DatabasePort, "database-port", "5432", "Database port")

	frameworkConf := &config.FrameworkConfig{}
	fs.BoolVar(&frameworkConf.IsTestBuild, "test-build", false, "run test environment and using test database")
	fs.StringVar(&frameworkConf.ServerPort, "port", "8080", "server serves and listens at port")

	if err := fs.Parse(args); err != nil {
		return err
	}

	a.Conf.DataSource = dataSource
	a.Conf.Framework = frameworkConf
	return nil
}
