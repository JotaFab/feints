// internal/commands/play.go
package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"feints/internal/core"
	"feints/internal/infra"
)

// PlayCommand reproduce o añade una canción a la cola
func PlayCommand(dp *infra.DiscordPlayer, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Obtener argumento (canción / búsqueda)
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ No se proporcionó ninguna canción.",
			},
		})
		return
	}

	query := options[0].StringValue()
	if query == "" {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ La búsqueda no puede estar vacía.",
			},
		})
		return
	}

	// Añadir canción a la cola
	dp.Add(core.Song{
		Title: query, // se puede enriquecer con metadatos de yt-dlp si quieres
		URL:   query,
	})
	
	

	// Responder al usuario
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("🎶 Añadido a la cola: **%s**", query),
		},
	})
}
