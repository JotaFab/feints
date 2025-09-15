package discord

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func StartBot(token string) error {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("error creating session: %w", err)
	}

	dg.AddHandler(interactionHandler)

	err = dg.Open()
	if err != nil {
		return fmt.Errorf("error opening connection: %w", err)
	}

	defer dg.Close()

	appID := dg.State.User.ID
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "play",
			Description: "Play a song",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "song",
					Description: "Song name or URL",
					Required:    true,
				},
			},
		},
		{Name: "stop", Description: "Stop playback"},
		{Name: "queue", Description: "Show the queue"},
		{Name: "next", Description: "Skip to next song"},
		{Name: "clear", Description: "Clear the queue"},
		{Name: "status", Description: "Show bot status"},
	}
	fmt.Println("🤖 Bot is now running. Press CTRL+C to exit.")
	for _, v := range commands {
		_, err := dg.ApplicationCommandCreate(appID, "", v)
		if err != nil {
			fmt.Printf("cannot create '%v' command: %v\n", v.Name, err)
		}
	}

	return waitForExit()
}

func waitForExit() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop
	fmt.Println("👋 Bot shutting down.")
	return nil
}
