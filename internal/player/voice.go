package player

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// connectVoice se encarga de unirse a un canal de voz de Discord.
func connectVoice(s *discordgo.Session, guildID, channelID string) (*discordgo.VoiceConnection, error) {
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return nil, err
	}
	log.Println("Successfully connected to voice channel.")
	return vc, nil
}

// disconnectVoice cierra la conexión de voz de forma segura.
func disconnectVoice(vc *discordgo.VoiceConnection) {
	if vc != nil {
		vc.Speaking(false)
		vc.Disconnect()
		log.Println("Disconnected from voice channel.")
	}
}
