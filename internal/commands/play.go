// internal/commands/play.go
package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"

	"feints/internal/player"
	"feints/internal/queue"
)

var Players = make(map[string]*player.Player) // un player por guild

// PlayCommand reproduce o añade una canción a la cola
func PlayCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	
	userID := i.Member.User.ID
	guildID := i.GuildID

	// buscar canal de voz del usuario
	var voiceChannelID string
	guild, err := s.State.Guild(guildID)
	if err != nil {
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

	// crear player si no existe
	p, ok := Players[guildID]
	if !ok {
		p, err = player.NewPlayer(s, guildID, voiceChannelID)
		if err != nil {
			log.Errorf("Error creando player: %v", err)
			return
		}
		Players[guildID] = p
	}

	// obtener argumento (canción / búsqueda)
	query := i.ApplicationCommandData().Options[0].StringValue()
	if query == "" {
		return
	}

	// añadir canción
	p.PlaySong(queue.Song{
		Title: query, // podrías enriquecer con yt-dlp metadata
		URL:   query,
	})

	// responder
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("🎶 Añadido a la cola: **%s**", query),
		},
	})
}
