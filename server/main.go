package main

import (
	"github.com/nhht77/earth-rest-api/pkg/logging"
	"github.com/nhht77/earth-rest-api/server/app"
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
