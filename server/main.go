package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/nhht77/earth-rest-api/pkg/logging"
	"github.com/nhht77/earth-rest-api/server/app"
)

func main() {
	log := logging.NewLogger()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go waitForSignal(ctx, cancel, log)

	a := app.New(log)

	if err := a.Setup(ctx); err != nil {
		log.Fatalf("Setup error: %s", err.Error())
	}
	defer a.Close()

	if err := a.Run(ctx); err != nil {
		cancel()
		log.Fatalf("Run error: %s", err.Error())
	}
}

func waitForSignal(ctx context.Context, cancel context.CancelFunc, log logging.Logger) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-signals:
		log.Infof("Received signal: %s", sig.String())
		cancel()
	case <-ctx.Done():
		return
	}
}
