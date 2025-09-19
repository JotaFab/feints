package botserver

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

func StartJanitor(bot *discordgo.Session) {
		CleanUpVoiceConnections(bot)
		time.Sleep(30 * time.Second)
}

func CleanUpVoiceConnections(bot *discordgo.Session) {
	for _, vc := range bot.VoiceConnections {
		vc.Disconnect()
	}
}