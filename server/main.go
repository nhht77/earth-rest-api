package main

import (
	"github.com/nhht77/earth-rest-api/server/app"
	"github.com/nhht77/earth-rest-api/server/pkg/logging"
)

func main() {
	log := logging.NewLogger()
	a := app.New(log)

	if err := a.Setup(); err != nil {
		log.Fatalf("Setup error: %s", err.Error())
	}
	defer a.Close()

	if err := a.Run(); err != nil {
		log.Fatalf("Run error: %s", err.Error())
	}
}
