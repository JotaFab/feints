package commands

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// runYtDlpSearch ejecuta yt-dlp con una búsqueda y devuelve títulos y URLs
func runYtDlpSearch(query string, max int) ([]string, []string, error) {
	// yt-dlp "ytsearch5:bohemian rhapsody" --get-title --get-id
	arg := fmt.Sprintf("ytsearch%d:%s", max, query)
	cmd := exec.Command("yt-dlp", arg, "--get-title", "--get-id")

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, nil, err
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines)%2 != 0 {
		return nil, nil, fmt.Errorf("unexpected yt-dlp output format")
	}

	titles := []string{}
	urls := []string{}
	for i := 0; i < len(lines); i += 2 {
		titles = append(titles, lines[i])
		urls = append(urls, "https://www.youtube.com/watch?v="+lines[i+1])
	}

	return titles, urls, nil
}

// SearchCmd maneja el autocompletado en /play search
func SearchCmd(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Solo manejar si es autocompletado
	if i.Type != discordgo.InteractionApplicationCommandAutocomplete {
		return
	}

	// Extraer query
	var query string
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == "search" {
			query = opt.StringValue()
		}
	}

	if query == "" {
		return
	}

	titles, urls, err := runYtDlpSearch(query, 5)
	log.Print(titles)
	if err != nil {
		fmt.Println("yt-dlp error:", err)
		return
	}

	choices := []*discordgo.ApplicationCommandOptionChoice{}
	for idx, title := range titles {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  title,
			Value: urls[idx], // aquí devolvemos el link como valor
		})
	}

	// Responder con sugerencias
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
	if err != nil {
		fmt.Println("autocomplete respond error:", err)
	}
}
