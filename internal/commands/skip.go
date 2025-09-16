package commands

import (
	"github.com/bwmarrin/discordgo"

	"feints/internal/player"
)

func SkipCommand(dp *player.DiscordPlayer, s *discordgo.Session, i *discordgo.InteractionCreate) {
	if err := dp.Skip(); err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Error saltando la canción.",
			},
		})
		return
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "⏭ Canción actual saltada.",
		},
	})
}
