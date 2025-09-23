// cmd/main.go
package main

import (

	"feints/internal/botserver"

	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func main() {
	// Ejecutar botserver en una goroutine
	
	log := loggerSetup()
	go func() {

		if err := botserver.Run(log); err != nil {
			log.Error("Bot server failed: %s", err)
			return
		}
	}()

	// Mantener el main vivo
	select {}
}



func loggerSetup() *slog.Logger {
	w := os.Stderr

	// Create a new logger
	logger := slog.New(tint.NewHandler(w, nil))

	// Set global logger with custom options
	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	return logger
}
