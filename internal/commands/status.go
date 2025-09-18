package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"feints/internal/player"
)

func StatusCommand(dp player.Player, s *discordgo.Session, i *discordgo.InteractionCreate) {
	status := dp.Status()
	current := dp.NowPlaying()
	msg := fmt.Sprintf("Estado: **%s**", status)
	if current.Title != "" {
		msg += fmt.Sprintf("\nReproduciendo: **%s**", current.Title)
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	})
}
