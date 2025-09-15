package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

// GetStatusMessage returns a status string for the bot
func GetStatusMessage(s *discordgo.Session) string {
	user := s.State.User
	guilds := len(s.State.Guilds)
	return fmt.Sprintf("ℹ️ Bot: %s\nServers: %d\nPing: %dms", user.Username, guilds, s.HeartbeatLatency().Milliseconds())
}
