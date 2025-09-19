package commands

import (
	"github.com/bwmarrin/discordgo"

	"feints/internal/core"
)

func StatusCommand(dp core.Player, s *discordgo.Session, i *discordgo.InteractionCreate) {

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: " status recived TODO",
		},
	})
}
