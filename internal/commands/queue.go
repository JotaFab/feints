package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"feints/internal/player"
)

func QueueCommand(dp player.Player, s *discordgo.Session, i *discordgo.InteractionCreate) {
	queue := dp.QueueList()
	if len(queue) == 0 {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "📭 La cola está vacía.",
			},
		})
		return
	}

	var sb strings.Builder
	for idx, song := range queue {
		sb.WriteString(fmt.Sprintf("%d. %s\n", idx+1, song.Title))
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("🎶 Cola actual:\n%s", sb.String()),
		},
	})
}
