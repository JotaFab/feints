package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/kkdai/youtube/v2"
	log "github.com/sirupsen/logrus"

	"feints/internal/player"
	"feints/internal/queue"
)

// TestCommand fills the queue with real YouTube songs
func TestCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Obtener el usuario que ejecuta el comando
	userID := i.Member.User.ID
	guildID := i.GuildID

	// Buscar el canal de voz donde está el usuario
	var voiceChannelID string
	guild, err := s.State.Guild(guildID)
	if err != nil {
		log.Errorf("Error obteniendo guild: %v", err)
		return
	}
	for _, vs := range guild.VoiceStates {
		if vs.UserID == userID {
			voiceChannelID = vs.ChannelID
			break
		}
	}
	if voiceChannelID == "" {
		log.Error("Usuario no está en ningún canal de voz")
		return
	}

	
	// crear player si no existe
	p, ok := Players[guildID]
	if !ok {
		p, err = player.NewPlayer(s, guildID, voiceChannelID)
		if err != nil {
			log.Errorf("Error creando player: %v", err)
			return
		}
		Players[guildID] = p
	}
	client := youtube.Client{}

	// Lista de canciones de prueba
	testSongs := []string{
		"https://www.youtube.com/watch?v=6JCLY0Rlx6Q", // Milky Chance - Colorado
		"https://www.youtube.com/watch?v=7wtfhZwyrcc", // Imagine Dragons - Believer
		"https://www.youtube.com/watch?v=fJ9rUzIMcZQ", // Queen - Bohemian Rhapsody
		"https://www.youtube.com/watch?v=kXYiU_JCYtU", // Linkin Park - Numb
		"https://www.youtube.com/watch?v=yKNxeF4KMsY", // Coldplay - Yellow
	}

	for _, url := range testSongs {
		video, err := client.GetVideo(url)
		if err != nil {
			log.Warnf("Error obteniendo info de %s: %v", url, err)
			continue
		}

		// Aquí usamos directamente context.Song
		p.PlaySong(queue.Song{
			Title: video.Title,
			URL:   url,
		})

	

		log.Infof("Encolada: %s | Autor: %s | %d seg", video.Title, video.Author, int(video.Duration.Seconds()))
	}

	// Responder en Discord
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("✅ Se añadieron %d canciones de prueba a la cola.", len(testSongs)),
		},
	})
}
