package commands

import (
	"encoding/json"
	"log"
	"os/exec"

	"github.com/bwmarrin/discordgo"
)

// SearchCommand maneja el autocompletado de /play search
func SearchCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	query := i.ApplicationCommandData().Options[0].StringValue()
	log.Println("[SearchCommand] Query recibida:", query)
	if query == "" {
		log.Println("[SearchCommand] Query vacía, saliendo")
		return
	}

	// Ejecutar yt-dlp para obtener resultados en JSON
	cmd := exec.Command("yt-dlp", "--dump-json", "--flat-playlist", "ytsearch5:"+query)
	out, err := cmd.Output()
	if err != nil {
		log.Println("[SearchCommand] Error ejecutando yt-dlp:", err)
		return
	}
	log.Println("[SearchCommand] yt-dlp ejecutado correctamente")

	// Cada línea es un JSON de resultado
	lines := string(out)
	var choices []*discordgo.ApplicationCommandOptionChoice
	for idx, line := range splitLines(lines) {
		if idx >= 25 {
			break
		}

		var video struct {
			Title      string `json:"title"`
			WebpageURL string `json:"url"`
		}
		if err := json.Unmarshal([]byte(line), &video); err != nil {
			log.Println("[SearchCommand] Error parseando línea JSON:", err)
			continue
		}

		name := video.Title
		if len(name) > 100 {
			name = name[:100]
		}
		log.Printf("[SearchCommand] Añadiendo choice: %s -> %s\n", name, video.WebpageURL)
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  name,
			Value: video.WebpageURL,
		})
	}

	if len(choices) == 0 {
		log.Println("[SearchCommand] No se encontraron resultados")
	}

	// Enviar resultados al usuario
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
	if err != nil {
		log.Println("[SearchCommand] Error enviando autocompletado:", err)
	} else {
		log.Println("[SearchCommand] Autocompletado enviado correctamente")
	}
}

// splitLines separa cada línea de salida de yt-dlp
func splitLines(s string) []string {
	var res []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			if start < i {
				res = append(res, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		res = append(res, s[start:])
	}
	return res
}
