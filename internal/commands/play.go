package commands

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"

	"layeh.com/gopus"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"

	"strings"
)

// sendTestAudio envía ruido blanco por OpusSend para pruebas

// sendTestAudio envía ruido blanco por OpusSend para pruebas
func sendTestAudio(vc *discordgo.VoiceConnection, durationSeconds int) error {
	err := vc.Speaking(true)
	if err != nil {
		return fmt.Errorf("failed to set speaking: %w", err)
	}
	defer vc.Speaking(false)
	const (
		sampleRate   = 48000
		channels     = 2
		frameSize    = 960
		pcmFrameSize = frameSize * channels * 2
	)
	encoder, err := gopus.NewEncoder(sampleRate, channels, gopus.Audio)
	if err != nil {
		return fmt.Errorf("failed to create opus encoder: %w", err)
	}
	log.WithField("duration", durationSeconds).Info("Sending test audio")
	frames := sampleRate * durationSeconds / frameSize
	start := time.Now()
	for f := 0; f < frames; f++ {
		pcm16 := make([]int16, frameSize*channels)
		for i := range pcm16 {
			pcm16[i] = int16((f*31+i*17)%65536 - 32768)
		}
		encoded, err := encoder.Encode(pcm16, frameSize, pcmFrameSize)
		if err != nil {
			return fmt.Errorf("failed to encode opus: %w", err)
		}
		select {
		case vc.OpusSend <- encoded:
		default:
		}
		time.Sleep(20 * time.Millisecond)
	}
	elapsed := time.Since(start)
	if elapsed < time.Duration(durationSeconds)*time.Second {
		log.WithFields(log.Fields{
			"expected": durationSeconds,
			"actual":   elapsed.Seconds(),
		}).Warn("Test audio finished earlier than expected")
	} 
	return nil
}

var (
	ErrNoSongSpecified   = errors.New("no song specified")
	ErrNoVoiceConnection = errors.New("no voice connection")
)

// PlayCommand handles the /play command, connecting to the user's voice channel
func PlayCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := "unknown"
	if i.Member != nil && i.Member.User != nil {
		user = i.Member.User.Username
	}
	log.WithFields(log.Fields{
		"user":    user,
		"command": "play",
	}).Info("Client action received")

	guildID := i.GuildID
	userID := i.Member.User.ID

	// Find the voice channel the user is in
	var voiceChannelID string
	guild, err := s.State.Guild(guildID)
	if err != nil {
		log.WithError(err).Error("Failed to get guild from state")
		Reply(s, i, "❌ Could not find guild info")
		return
	}
	log.WithField("guild", guild).Debug("Guild info")
	for _, vs := range guild.VoiceStates {
		if vs.UserID == userID {
			voiceChannelID = vs.ChannelID
			break
		}
	}
	if voiceChannelID == "" {
		Reply(s, i, "❌ You must be in a voice channel to use /play")
		return
	}

	// Connect to the voice channel
	vc, err := s.ChannelVoiceJoin(guildID, voiceChannelID, false, true)
	if err != nil {
		Reply(s, i, "❌ Failed to join voice channel")
		return
	}
	song := i.ApplicationCommandData().Options[0].StringValue()
	log.WithField("song", song).Info("Requested song")
	// Prueba de audio: envía ruido blanco por 3 segundos
	err = sendTestAudio(vc, 300)
	if err != nil {
		Reply(s, i, "❌ Test audio failed: "+err.Error())
		return
	}
	log.WithFields(log.Fields{
		"user": user,
		"test": true,
	}).Info("Test audio sent")

	// ...lógica original de reproducción...
	// err = PlaySong(vc, song)
	// if err != nil {
	//     Reply(s, i, "❌ Failed to play the song: "+err.Error())
	//     return
	// }
	// log.WithFields(log.Fields{
	//     "user": user,
	//     "song": song,
	// }).Info("Playing song")

	Reply(s, i, "✅ Test audio enviado al canal de voz!")
}

// PlaySong plays a song in a Discord voice channel using yt-dlp and ffmpeg.
func PlaySong(vc *discordgo.VoiceConnection, song string) error {
	url, err := searchSongURL(song)
	if err != nil {
		return err
	}
	ffmpegOut, ffmpegCmd, err := getAudioStream(url)
	if err != nil {
		return err
	}
	err = sendAudioStream(vc, ffmpegOut, ffmpegCmd)
	if err != nil {
		return err
	}
	return nil
}

// searchSongURL busca la URL de streaming usando yt-dlp
func searchSongURL(song string) (string, error) {
	urlCmd := exec.Command("yt-dlp", "--get-url", "ytsearch:"+song)
	urlOutput, err := urlCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get song URL: %w", err)
	}
	url := strings.TrimSpace(string(urlOutput))
	log.WithFields(log.Fields{
		"song": song,
		"url":  url,
	}).Info("Found song URL")
	return url, nil
}

// getAudioStream inicia ffmpeg y retorna el pipe de audio y el comando
func getAudioStream(url string) (io.ReadCloser, *exec.Cmd, error) {
	ffmpegCmd := exec.Command("ffmpeg", "-i", url, "-f", "s16le", "-ar", "48000", "-ac", "2", "pipe:1")
	ffmpegOut, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get ffmpeg stdout pipe: %w", err)
	}
	if err := ffmpegCmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("failed to start ffmpeg: %w", err)
	}
	return ffmpegOut, ffmpegCmd, nil
}

// sendAudioStream codifica y envía el audio PCM por OpusSend
func sendAudioStream(vc *discordgo.VoiceConnection, ffmpegOut io.ReadCloser, ffmpegCmd *exec.Cmd) error {
	vc.Speaking(true)
	defer vc.Speaking(false)
	const (
		sampleRate   = 48000
		channels     = 2
		frameSize    = 960
		pcmFrameSize = frameSize * channels * 2
	)
	encoder, err := gopus.NewEncoder(sampleRate, channels, gopus.Audio)
	if err != nil {
		return fmt.Errorf("failed to create opus encoder: %w", err)
	}
	pcmBuf := make([]byte, pcmFrameSize)
	for {
		_, err := io.ReadFull(ffmpegOut, pcmBuf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read PCM from ffmpeg: %w", err)
		}
		pcm16 := make([]int16, frameSize*channels)
		for i := 0; i < len(pcm16); i++ {
			pcm16[i] = int16(pcmBuf[i*2]) | int16(pcmBuf[i*2+1])<<8
		}
		encoded, err := encoder.Encode(pcm16, frameSize, pcmFrameSize)
		if err != nil {
			return fmt.Errorf("failed to encode opus: %w", err)
		}
		select {
		case vc.OpusSend <- encoded:
		default:
		}
	}
	if err := ffmpegCmd.Wait(); err != nil && err.Error() != "signal: killed" {
		return fmt.Errorf("ffmpeg process exited with an error: %w", err)
	}
	return nil
}
