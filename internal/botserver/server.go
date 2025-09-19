package botserver

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"

	"feints/internal/commands"
    "feints/internal/infra"

	_ "github.com/davecgh/go-spew/spew"
)

// BotServer administra múltiples reproductores por guild
type BotServer struct {
	session       *discordgo.Session
	log *log.Logger
}

// NewBotServer crea un nuevo servidor de bots
func NewBotServer(s *discordgo.Session) *BotServer {
logger := log.New(os.Stdout, "[BOT-server] ", log.LstdFlags)

	return &BotServer{
		session:       s,
		log: logger,
	}
}

func (bs *BotServer) GetOrCreatePlayer(guildID, channelID string) (*infra.DiscordPlayer, error) {

    // Infraestructura: decorador que lo conecta a Discord
	dp, err := infra.NewDiscordPlayer(bs.session, guildID, channelID, bs.log)
	if err != nil {
		return nil, err
	}


    return dp, nil
}


// HandleCommand despacha las interacciones a los comandos
func (bs *BotServer) HandleCommand(cmd string, s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID
	guildID := i.GuildID

	// Buscar canal de voz del usuario
	var voiceChannelID string
	guild, err := s.State.Guild(guildID)
	if err != nil {
		log.Println("No se pudo obtener guild:", err)
		return
	}
	for _, vs := range guild.VoiceStates {
		if vs.UserID == userID {
			voiceChannelID = vs.ChannelID
			break
		}
	}
	if voiceChannelID == "" {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ No estás en un canal de voz.",
			},
		})
		return
	}

	// Obtener el player correspondiente
	dp, err := bs.GetOrCreatePlayer(guildID, voiceChannelID)
	if err != nil {
		log.Println("Error obteniendo player:", err)
		return
	}
	bs.log.Println(cmd)

	// Enviar el player al comando
	switch cmd {
	case "play":
		commands.PlayCommand(dp, s, i) // ahora PlayCommand recibe solo su player
	case "pause":
		// commands.PauseCommand(dp, s, i)
	case "stop":
		commands.StopCommand(dp, s, i)
	case "queue":
		commands.QueueCommand(dp, s, i)
	case "skip" :
		commands.SkipCommand(dp, s, i)
	case "next":
		commands.SkipCommand(dp, s, i)
	case "clear":
		commands.ClearCommand(dp, s, i)
	case "status":
		commands.StatusCommand(dp, s, i)
	case "test":
		commands.TestCommand(dp, s, i)
		//commands.TestMultiVoiceCommand(bs.PlayerManager, s,i)
	}
}

// Run inicializa el bot y maneja los eventos
func Run() error {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		return fmt.Errorf("DISCORD_BOT_TOKEN no está definido")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("error creando sesión de Discord: %v", err)
	}

	bs := NewBotServer(dg)

	// Al estar listo el bot
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("✅ Bot conectado como", s.State.User.Username)

		commandsToRegister := []*discordgo.ApplicationCommand{
			{
				Name:        "play",
				Description: "Reproduce una canción",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "search",
						Description: "Nombre o URL de la canción",
						Required:    true,
						Autocomplete: true,
					},
				},
			},
			{Name: "stop", Description: "Detiene la reproducción y se desconecta"},
			{Name: "queue", Description: "Muestra la cola de canciones"},
			{Name: "skip", Description: "Salta a la siguiente canción"},
			{Name: "clear", Description: "Limpia la cola"},
			{Name: "status", Description: "Muestra el estado actual"},
			{Name: "test", Description: "Prueba de carga en la cola"},
		}

		for _, cmd := range commandsToRegister {
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
			log.Printf("how many times i was here?? %s", i.Locale.String())
			switch i.ApplicationCommandData().Name {
			case "play":
				commands.SearchCommand(s, i) // llama la función de autocompletado
			}
		}
	})

	// Registrar handler de comandos
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		if i.Type == discordgo.InteractionApplicationCommand {
			bs.HandleCommand(i.ApplicationCommandData().Name, s, i)
		}
	})

	if err := dg.Open(); err != nil {
		return fmt.Errorf("error abriendo conexión de Discord: %v", err)
	}
	go StartJanitor(dg)
	defer dg.Close()

	fmt.Println("🤖 Bot is running. Press CTRL+C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	return nil
}


