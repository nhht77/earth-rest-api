package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

var (
	Log = logrus.New()

	DB        = &Database{}
	AppConfig = &Config{}
)

func init_resource() {

	// init logger
	init_logger()

	AppConfig.ReadDefault()

	if err := DB.Initialize(AppConfig); err != nil {
		release_resource()
		Log.Fatalf("Error: open connection %s", err.Error())
		return
	}
}

func main() {
	init_resource()
}

func release_resource() {
	DB.Close()
}

////// Logger
func init_logger() {
	Log.Out = os.Stdout
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}
