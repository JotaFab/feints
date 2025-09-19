package commands

import (
	"github.com/bwmarrin/discordgo"

	"feints/internal/core"
)

func SkipCommand(dp core.Player, s *discordgo.Session, i *discordgo.InteractionCreate) {
	dp.Next()
	

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "⏭ Ahora reproduciendo ",
		},
	})
}
