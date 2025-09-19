package commands

import (
	"github.com/bwmarrin/discordgo"

	"feints/internal/infra"
)

func StopCommand(dp *infra.DiscordPlayer, s *discordgo.Session, i *discordgo.InteractionCreate) {
	dp.Stop()

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "⏹ Reproducción detenida y cola limpiada.",
		},
	})
}

func ClearCommand(dp *infra.DiscordPlayer, s *discordgo.Session, i *discordgo.InteractionCreate) {
	dp.Stop()

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "⏹ cola limpiada.",
		},
	})
}
