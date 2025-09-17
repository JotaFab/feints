package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"feints/internal/player"
)

// SearchCommand maneja el autocompletado de /play search
func SearchCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	query := i.ApplicationCommandData().Options[0].StringValue()
	log.Println("[SearchCommand] Query recibida:", query)
	if query == "" {
		log.Println("[SearchCommand] Query vacía, saliendo")
		return
	}

	// Llamamos a nuestro wrapper YtdlpSearch
	results, err := player.YtdlpSearch(query, 10)
	if err != nil {
		log.Println("[SearchCommand] Error ejecutando yt-dlp:", err)
		return
	}
	log.Println("[SearchCommand] yt-dlp ejecutado correctamente")

	var choices []*discordgo.ApplicationCommandOptionChoice
	for idx, line := range results {
		if idx >= 25 {
			break
		}

		var video struct {
			Title      string  `json:"title"`
			WebpageURL string  `json:"webpage_url"` // 👈 aquí usa el campo correcto
			Duration   float64 `json:"duration"`
			IsLive     bool    `json:"is_live"`
		}

		if err := json.Unmarshal([]byte(line), &video); err != nil {
			log.Println("[SearchCommand] Error parseando línea JSON:", err)
			continue
		}

		if video.IsLive || video.Duration > 900 {
			log.Printf("[SearchCommand] Descartando video: '%s' (live o >15m)\n", video.Title)
			continue
		}

		// Duración bonita
		durationFormatted := formatDuration(video.Duration)

		name := fmt.Sprintf("[%s] %s", durationFormatted, video.Title)
		if len(name) > 100 {
			name = name[:100]
		}
		log.Printf("[SearchCommand] Añadiendo choice: %s -> %s\n", name, video.WebpageURL)

		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  name,
			Value: video.WebpageURL, // 👈 esto es lo que luego usarás en YtdlpBestAudioURL
		})
	}

	if len(choices) == 0 {
		log.Println("[SearchCommand] No se encontraron resultados válidos")
	}

	// Responder al usuario
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


// formatDuration convierte segundos en un string con formato MM:SS
func formatDuration(duration float64) string {
	d := time.Duration(duration) * time.Second
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}