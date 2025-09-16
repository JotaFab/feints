package commands

import (
	"github.com/bwmarrin/discordgo"

	"feints/internal/player"
)

func StopCommand(dp *player.DiscordPlayer, s *discordgo.Session, i *discordgo.InteractionCreate) {
	if err := dp.Skip(); err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Error deteniendo la canción.",
			},
		})
		return
	}

	dp.Clear()

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "⏹ Reproducción detenida y cola limpiada.",
		},
	})
}
