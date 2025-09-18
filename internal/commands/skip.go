package commands

import (
	"github.com/bwmarrin/discordgo"

	"feints/internal/player"
)

func SkipCommand(dp player.Player, s *discordgo.Session, i *discordgo.InteractionCreate) {
	dp.NextSong()
	current := dp.NowPlaying()
	currentTitle := ""
	if current.Title != "" {
		currentTitle = current.Title
	}
	dp.Play()

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "⏭ Ahora reproduciendo: " + currentTitle,
		},
	})
}
