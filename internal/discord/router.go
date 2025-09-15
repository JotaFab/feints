package discord

import (
	"feints/internal/commands"

	"github.com/bwmarrin/discordgo"
)

// ==== INTERACTION ROUTER ====
func interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	var cmd commands.Command
	switch i.ApplicationCommandData().Name {
	case "play":
		cmd = commands.CmdPlay
	case "stop":
		cmd = commands.CmdStop
	case "queue":
		cmd = commands.CmdQueue
	case "next":
		cmd = commands.CmdNext
	case "clear":
		cmd = commands.CmdClear
	case "status":
		cmd = commands.CmdStatus
	default:
		return
	}

	commands.HandleCommand(cmd, s, i)
}
