package commands

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
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
		PlayCommand(s, i)
	case CmdStop:
		Reply(s, i, "⏹ Stop command received")
	case CmdQueue:
		Reply(s, i, "📜 Queue command received")
	case CmdNext:
		Reply(s, i, "⏭ Next command received")
	case CmdClear:
		Reply(s, i, "🧹 Clear command received")
	case CmdStatus:
		Reply(s, i, GetStatusMessage(s))
	}
}

func Reply(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: msg},
	})
}
