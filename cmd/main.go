// cmd/main.go
package main

import (
	"log"

	"feints/internal/botserver"
)

func main() {
	// Ejecutar botserver en una goroutine
	go func() {
		if err := botserver.Run(); err != nil {
			log.Fatalf("Bot server failed: %v", err)
		}
	}()

	// Mantener el main vivo
	select {}
}
