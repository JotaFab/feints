package commands

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"

	"feints/internal/infra"
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
	results, err := infra.YtdlpSearch(query, 10)
	if err != nil {
		log.Println("[SearchCommand] Error ejecutando yt-dlp:", err)
		return
	}
	log.Println("[SearchCommand] yt-dlp ejecutado correctamente")

	var choices []*discordgo.ApplicationCommandOptionChoice
	for idx, video := range results {
		if idx >= 25 {
			break
		}

		name := fmt.Sprintf("[%s] %s", video.Duration, video.Title)
		if len(name) > 100 {
			name = name[:100]
		}
		log.Printf("[SearchCommand] Añadiendo choice: %s -> %s\n", name, video.URL)

		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  name,
			Value: video.URL, // 👈 esto es lo que luego usarás en YtdlpBestAudioURL
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


