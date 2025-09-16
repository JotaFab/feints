// internal/commands/skip.go
package commands

import (
	"github.com/bwmarrin/discordgo"
)

func SkipCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID
	if p, ok := Players[guildID]; ok {
		p.Skip()
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "⏭️ Canción saltada.",
			},
		})
	}
}
