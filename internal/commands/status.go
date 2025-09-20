package commands

import (
	"github.com/bwmarrin/discordgo"

	"feints/internal/core"
	"fmt"
)

func StatusCommand(dp core.Player, s *discordgo.Session, i *discordgo.InteractionCreate) {
	var state string
	state = dp.State()
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf(" status: %s", state),
		},
	})
}
