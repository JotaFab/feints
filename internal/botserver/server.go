// internal/botserver/botserver.go
package botserver

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Run inicializa el bot y maneja los eventos
func Run() error {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		return fmt.Errorf("DISCORD_BOT_TOKEN no está definido")
	}

	// Crear sesión de Discord
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("error creando sesión de Discord: %v", err)
	}

	// Registrar handler de interacciones
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Aquí deberíamos mapear el nombre del comando a nuestro Command enum
		switch i.ApplicationCommandData().Name {
		case "play":
			HandleCommand(CmdPlay, s, i)
		case "stop":
			HandleCommand(CmdStop, s, i)
		case "queue":
			HandleCommand(CmdQueue, s, i)
		case "next":
			HandleCommand(CmdNext, s, i)
		case "clear":
			HandleCommand(CmdClear, s, i)
		case "status":
			HandleCommand(CmdStatus, s, i)
		}
	})

	// Abrir conexión WebSocket con Discord
	if err := dg.Open(); err != nil {
		return fmt.Errorf("error abriendo conexión de Discord: %v", err)
	}
	defer dg.Close()

	// Esperar a que termine con Ctrl+C
	fmt.Println("🤖 Bot is running. Press CTRL+C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	return nil
}
