package botserver

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"feints/internal/commands"
	"feints/internal/infra"
	"feints/internal/core"

	"github.com/bwmarrin/discordgo"
	"log/slog"
)

// BotServer administra múltiples reproductores por guild
type BotServer struct {
	session *discordgo.Session
	Log     *slog.Logger
	players	map[string]core.Player
}

// NewBotServer crea un nuevo servidor de bots con logger JSON
func NewBotServer(s *discordgo.Session, logger *slog.Logger) *BotServer {

	return &BotServer{
		session: s,
		Log:     logger,
		players: make(map[string]core.Player),
	}
}

func (bs *BotServer) GetOrCreatePlayer(guildID, channelID string) (core.Player, error) {
    key := guildID + "_" + channelID

    // si ya existe, devolverlo
    if player, ok := bs.players[key]; ok {
        bs.Log.Info("Player encontrado", "guildID", guildID, "channelID", channelID)
        return player, nil
    }

    // crear uno nuevo
    dp := infra.NewDgvoicePlayer(bs.session, guildID, channelID, bs.Log)
    bs.players[key] = dp

    bs.Log.Info("Player creado", "guildID", guildID, "channelID", channelID)
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
		bs.Log.Error("No se pudo obtener guild", "err", err, "guildID", guildID)
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

	// Obtener el player
	dp, err := bs.GetOrCreatePlayer(guildID, voiceChannelID)
	if err != nil {
		bs.Log.Error("Error obteniendo player", "err", err)
		return
	}

	bs.Log.Info("Ejecutando comando", "cmd", cmd, "userID", userID, "guildID", guildID)

	switch cmd {
	case "play":
		commands.PlayCommand(dp, s, i)
	case "pause":
		// commands.PauseCommand(dp, s, i)
	case "stop":
		commands.StopCommand(dp, s, i)
	case "queue":
		commands.QueueCommand(dp, s, i)
	case "skip", "next":
		commands.SkipCommand(dp, s, i)
	case "clear":
		commands.ClearCommand(dp, s, i)
	case "status":
		commands.StatusCommand(dp, s, i)
	case "test":
		commands.TestCommand(dp, s, i)
	case "autoplay":
		commands.AutoPlay(dp,s,i)
	}
}

// Run inicializa el bot y maneja los eventos
func Run(log*slog.Logger) error {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		return fmt.Errorf("DISCORD_BOT_TOKEN no está definido")
	}
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("error creando sesión de Discord: %v", err)
	}
	
	log = log.With("component", "BotServer")

	// Handler de Ready
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Info("Bot conectado", "username", s.State.User.Username)

		commandsToRegister := []*discordgo.ApplicationCommand{
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
			{Name: "stop", Description: "Detiene la reproducción y se desconecta"},
			{Name: "queue", Description: "Muestra la cola de canciones"},
			{Name: "skip", Description: "Salta a la siguiente canción"},
			{Name: "clear", Description: "Limpia la cola"},
			{Name: "status", Description: "Muestra el estado actual"},
			{Name: "test", Description: "Prueba de carga en la cola"},
			{Name: "autoplay", Description: "Activa autoplay"},
		}

		for _, cmd := range commandsToRegister {
			_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
			if err != nil {
				log.Error("Error registrando comando", "name", cmd.Name, "err", err)
			} else {
				log.Info("Comando registrado", "name", cmd.Name)
			}
		}
	})

	// Autocompletado
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommandAutocomplete {
			switch i.ApplicationCommandData().Name {
			case "play":
				commands.SearchCommand(s, i)
			}
		}
	})
	if err := dg.Open(); err != nil {
		return err
	}
	bs := NewBotServer(dg, log)
	// Comandos
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			bs.HandleCommand(i.ApplicationCommandData().Name, s, i)
		}
	})

	StartJanitor(dg)
	defer dg.Close()

	bs.Log.Info("Bot ejecutándose. Presiona CTRL+C para salir.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	return nil
}
