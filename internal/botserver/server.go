// internal/botserver/botserver.go
package botserver

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"

	"feints/internal/commands"
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
	// Después de crear la sesión dg := discordgo.New("Bot " + token)
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("✅ Bot conectado como", s.State.User.Username)

		commands := []*discordgo.ApplicationCommand{
			{
				Name:        "play",
				Description: "Reproduce una canción",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:         discordgo.ApplicationCommandOptionString,
						Name:         "search",
						Description:  "Nombre o URL de la canción",
						Required:     true,
						Autocomplete: true,
					},
				},
			},
			{
				Name:        "stop",
				Description: "Detiene la reproducción y se desconecta",
			},
			{
				Name:        "queue",
				Description: "Muestra la cola de canciones",
			},
			{
				Name:        "skip",
				Description: "Salta a la siguiente canción",
			},
			{
				Name:        "clear",
				Description: "Limpia la cola",
			},
			{
				Name:        "status",
				Description: "Muestra el estado actual",
			},
			{
				Name:        "test",
				Description: "Prueba de carga en la cola",
			},
		}

		// Registrar todos los comandos en la aplicación (globales)
		for _, cmd := range commands {
			_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
			if err != nil {
				fmt.Printf("❌ Error registrando comando %s: %v\n", cmd.Name, err)
			} else {
				fmt.Printf("✅ Comando /%s registrado\n", cmd.Name)
			}
		}
	})
		// Handler para autocompletado
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommandAutocomplete {
			log.Println(i.ApplicationCommandData().Options[0].StringValue())
			switch i.ApplicationCommandData().Name {
			case "play":
				commands.SearchCmd(s, i)
			}
		}
	})
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
		case "skip":
			HandleCommand(CmdSkip, s, i)
		case "clear":
			HandleCommand(CmdClear, s, i)
		case "status":
			HandleCommand(CmdStatus, s, i)
		case "test":
			HandleCommand(CmdTest, s, i)
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
