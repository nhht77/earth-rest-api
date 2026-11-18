package app

import (
	"flag"

	"github.com/nhht77/earth-rest-api/server/config"
	"github.com/nhht77/earth-rest-api/server/database"
	"github.com/nhht77/earth-rest-api/server/handlers"
	"github.com/nhht77/earth-rest-api/server/pkg/logging"
	"github.com/nhht77/earth-rest-api/server/pkg/mstring"
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

func (a *App) Setup() error {
	a.initFramework() // your init_framework, method-ified: flags + Log.Info line
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
func (a *App) Run() error {
	return handlers.RunHTTP(a.Conf, a.Handlers)
}

// Close releases resources — called via defer in main, even on error paths.
func (a *App) Close() {
	a.DB.Close()
}

func (a *App) initFramework() {
	frameworkConf := &config.FrameworkConfig{}
	flag.BoolVar(&frameworkConf.IsTestBuild, "test-build", false, "run test environment and using test database")
	flag.StringVar(&frameworkConf.ServerPort, "port", "8080", "server serves and listens at port")
	flag.StringVar(&frameworkConf.DatabasePort, "database-port", "5432", "Database port")
	flag.StringVar(&frameworkConf.DatabaseHost, "database-host", "localhost", "Database host")
	flag.Parse()

	a.Conf.Framework = frameworkConf
	a.Log.Info("Framework Config: ", mstring.ToJSON(a.Conf.Framework))
}
