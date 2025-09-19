package commands

import (
	"github.com/bwmarrin/discordgo"

	"feints/internal/core"
)

func StopCommand(dp core.Player, s *discordgo.Session, i *discordgo.InteractionCreate) {
	dp.Stop()

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "⏹ Reproducción detenida y cola limpiada.",
		},
	})
}

func ClearCommand(dp core.Player, s *discordgo.Session, i *discordgo.InteractionCreate) {
	dp.Stop()

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "⏹ cola limpiada.",
		},
	})
}
