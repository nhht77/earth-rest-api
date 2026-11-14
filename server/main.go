package main

import (
	"flag"

	"github.com/nhht77/earth-rest-api/server/pkg/logging"
	"github.com/nhht77/earth-rest-api/server/pkg/mstring"
)

var (
	Log = logging.NewLogger()

	DB        = &Database{}
	AppConfig = &Config{}
)

func init_resource() {

	// init framework
	init_framework(AppConfig)

	AppConfig.ReadDefault()

	if err := DB.Initialize(AppConfig); err != nil {
		release_resource()
		Log.Fatalf("Error: open connection %s", err.Error())
		return
	}
}

func main() {
	init_resource()

	// Note: clean table after each resource release
	// @todo Temp solution until wrap application in docker container
	if AppConfig.Framework.IsTestBuild {
		DB._ClearTable()
	}

	if err := RunHTTP(); err != nil {
		release_resource()
		Log.Fatalf("[mhttp] RunHTTP error %s", err.Error())
		return
	}
}

func release_resource() {
	DB.Close()
}

func init_framework(AppConfig *Config) {

	AppConfig.Framework = &FrameworkConfig{}

	flag.BoolVar(&AppConfig.Framework.IsTestBuild, "-test-build", false, "run test environment and using test database")
	flag.StringVar(&AppConfig.Framework.ServerPort, "-port", "8080", "server serves and listens at port")
	flag.StringVar(&AppConfig.Framework.DatabasePort, "-database-port", "5432", "Database port")
	flag.StringVar(&AppConfig.Framework.DatabaseHost, "-database-host", "localhost", "Database host")

	Log.Info("Framework Config: ", mstring.ToJSON(AppConfig.Framework))
}
