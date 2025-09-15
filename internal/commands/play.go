// internal/commands/play.go
package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"

	"feints/internal/audio"
	"feints/internal/context"
)

// PlayCommand reproduce una canción en un canal de voz.
// i: interacción de Discord
func PlayCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	// Obtener el usuario que ejecuta el comando
	userID := i.Member.User.ID
	guildID := i.GuildID

	// Buscar el canal de voz donde está el usuario

	// Find the voice channel the user is in
	var voiceChannelID string
	guild, err := s.State.Guild(guildID)
	if err != nil {
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

	// Obtener la canción del comando
	song := i.ApplicationCommandData().Options[0].StringValue()
	if song == "" {
		log.Error("No se especificó ninguna canción")
		return
	}

	// Conectar al canal de voz y obtener contexto
	ctx, err := context.GetOrCreateContext(s, guildID, voiceChannelID)
	if err != nil {
		log.Errorf("Error conectando al canal de voz: %v", err)
		return
	}
	// Responder inmediatamente para que Discord no marque "La aplicación no ha respondido"
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Errorf("Error respondiendo interacción: %v", err)
		return
	}

	log.Infof("Buscando canción: %s", song)
	meta, err := audio.GetSongMetadata(song)
	if err == nil {
		log.Infof("Título: %s | Autor: %s | Duración: %d seg", meta.Title, meta.Uploader, meta.Duration)
	}
	message := fmt.Sprintf("🎵 **%s** \n Autor: **%s** \n Duración: %d seg \n Url: **%s**", meta.Title, meta.Uploader, meta.Duration, meta.WebpageURL)
	_, err = s.ChannelMessageSend(i.ChannelID, message)
	if err != nil {
		log.Errorf("Error enviando mensaje de Discord: %v", err)
	}
	// Reproducir la canción
	if err := PlaySong(ctx, song); err != nil {
		log.Errorf("Error reproduciendo la canción: %v", err)
	}
}

// PlaySong obtiene el stream y lo envía a Discord
func PlaySong(ctx *context.VoiceContext, song string) error {
	log.Infof("Buscando canción: %s", song)

	ffmpegOut, ffmpegCmd, err := audio.GetAudioStream(song)
	if err != nil {
		return fmt.Errorf("error obteniendo audio: %w", err)
	}

	// Enviar audio en goroutine
	go func() {
		if err := audio.SendAudioStream(ctx, ffmpegOut, ffmpegCmd); err != nil {
			log.Errorf("Error en reproducción: %v", err)
		}
	}()

	// Marcar que está reproduciendo
	ctx.Mutex.Lock()
	ctx.Playing = true
	log.Info("reproduciendo se supone")
	ctx.Mutex.Unlock()

	log.Infof("Reproduciendo: %s", strings.TrimSpace(song))
	return nil
}
