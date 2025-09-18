package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"

	"feints/internal/player"
)

func StopCommand(dp player.Player, s *discordgo.Session, i *discordgo.InteractionCreate) {
	dp.Stop()
	log.Println("entering stop ??? ", dp.Status())


	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "⏹ Reproducción detenida y cola limpiada.",
		},
	})
}

func ClearCommand(dp player.Player, s *discordgo.Session, i *discordgo.InteractionCreate) { 
	dp.Clear()
	log.Println("entering stop ??? ", dp.Status())


	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "⏹ cola limpiada.",
		},
	})
}