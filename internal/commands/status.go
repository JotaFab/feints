package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"feints/internal/player"
)

func StatusCommand(dp *player.DiscordPlayer, s *discordgo.Session, i *discordgo.InteractionCreate) {
	status := dp.Status()
	current := dp.Current()
	msg := fmt.Sprintf("Estado: **%s**", status)
	if current != nil {
		msg += fmt.Sprintf("\nReproduciendo: **%s**", current.Title)
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	})
}
