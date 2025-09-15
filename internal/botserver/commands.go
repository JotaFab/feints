// internal/botserver/commands.go
package botserver

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"

	"feints/internal/commands" // importa tu paquete commands de lógica
)

type Command int

const (
	CmdPlay Command = iota
	CmdStop
	CmdQueue
	CmdNext
	CmdClear
	CmdStatus
)

// HandleCommand recibe un Command y lo redirige a la función correspondiente en internal/commands
func HandleCommand(cmd Command, s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := "unknown"
	if i.Member != nil && i.Member.User != nil {
		user = i.Member.User.Username
	}
	log.WithFields(log.Fields{
		"user":    user,
		"command": cmd,
	}).Info("Client action received")

	switch cmd {
	case CmdPlay:
		commands.PlayCommand(s, i)
	case CmdStop:
		// commands.StopCommand(s, i)
	case CmdQueue:
		//commands.Reply(s, i, "📜 Queue command received")
	case CmdNext:
		//commands.Reply(s, i, "⏭ Next command received")
	case CmdClear:
		//commands.Reply(s, i, "🧹 Clear command received")
	case CmdStatus:
		//commands.Reply(s, i, commands.GetStatusMessage(s))
	}
}
