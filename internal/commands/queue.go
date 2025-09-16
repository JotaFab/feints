// internal/commands/queue.go
package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func QueueCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID
	if p, ok := Players[guildID]; ok {
		q := p.ShowQueue()
		if len(q) == 0 {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "📭 La cola está vacía.",
				},
			})
			return
		}

		resp := "🎶 **Cola actual:**\n"
		for i, sng := range q {
			resp += fmt.Sprintf("%d. %s\n", i+1, sng.Title)
		}

		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: resp,
			},
		})
	}
}
