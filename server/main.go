package main

import (
	"flag"

	"github.com/nhht77/earth-rest-api/server/config"
	"github.com/nhht77/earth-rest-api/server/database"
	"github.com/nhht77/earth-rest-api/server/handlers"
	"github.com/nhht77/earth-rest-api/server/pkg/logging"
	"github.com/nhht77/earth-rest-api/server/pkg/mstring"
)

var (
	Log = logging.NewLogger()

	db      = database.NewDatabase(Log)
	appConf = &config.Config{}
)

func init_resource() {

	// init framework
	init_framework(appConf)

	appConf.ReadDefault(*appConf, Log)

	if err := db.Initialize(appConf); err != nil {
		release_resource()
		Log.Fatalf("Error: open connection %s", err.Error())
		return
	}
}

func main() {
	init_resource()

	// Note: clean table after each resource release
	// @todo Temp solution until wrap application in docker container
	if appConf.Framework.IsTestBuild {
		db.ClearTable()
	}

	if err := handlers.RunHTTP(appConf, handlers.NewHandlers(db, Log)); err != nil {
		release_resource()
		Log.Fatalf("[mhttp] RunHTTP error %s", err.Error())
		return
	}
}

func release_resource() {
	db.Close()
}

func init_framework(appConf *config.Config) {

	appConf.Framework = &config.FrameworkConfig{}

	// flag names must NOT start with '-' — the flag package adds the dash
	// itself and panics on names that begin with one (Go >= 1.10).
	flag.BoolVar(&appConf.Framework.IsTestBuild, "test-build", false, "run test environment and using test database")
	flag.StringVar(&appConf.Framework.ServerPort, "port", "8080", "server serves and listens at port")
	flag.StringVar(&appConf.Framework.DatabasePort, "database-port", "5432", "Database port")
	flag.StringVar(&appConf.Framework.DatabaseHost, "database-host", "localhost", "Database host")

	// Parse the command line — without this, every registered flag keeps its
	// default value and CLI arguments are silently ignored.
	flag.Parse()

	Log.Info("Framework Config: ", mstring.ToJSON(appConf.Framework))
}
