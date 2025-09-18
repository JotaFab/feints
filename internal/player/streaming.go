package player

import (
	"encoding/binary"
	"io"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"layeh.com/gopus"
)

// StreamConfig contiene la configuración necesaria para el streaming.
type StreamConfig struct {
	VC          *discordgo.VoiceConnection
	Encoder     *gopus.Encoder
	OpusSender  chan []byte
	CurrentSong *Song
	StopChan    chan struct{}
	ControlChan chan controlMsg
	State       PlayerState
	Logger      *log.Logger
	StateMu     *sync.RWMutex
}

// StreamSongLoop maneja la lógica de streaming de audio.
func StreamSongLoop(cfg *StreamConfig) {
	encoder, err := gopus.NewEncoder(48000, 2, gopus.Audio)
	if err != nil {
		log.Println(err)
		return
	}
	cfg.Encoder = encoder
	audioURL, err := YtdlpBestAudioURL(cfg.CurrentSong.URL)
	if err != nil {
		cfg.Logger.Println("Error getting audio URL:", err)
		return
	}

	ffmpegCmd := exec.Command("ffmpeg", "-i", audioURL, "-f", "s16le", "-ar", "48000", "-ac", "2", "pipe:1")
	ffmpegOut, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		cfg.Logger.Println("FFmpeg pipe error:", err)
		return
	}
	if err := ffmpegCmd.Start(); err != nil {
		cfg.Logger.Println("FFmpeg start error:", err)
		return
	}

	const frameSize = 960
	buf := make([]byte, frameSize*2*2)

	for {

		cfg.StateMu.RLock()
		state := cfg.State
		cfg.StateMu.RUnlock()

		if state == Paused {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		_, err := io.ReadFull(ffmpegOut, buf)
		if err != nil {
			cfg.Logger.Println("End of audio stream or error:", err)
			ffmpegCmd.Process.Kill()
			cfg.ControlChan <- controlMsg{Command: "end_of_stream"}
			return
		}

		// Lógica de encoding y envío
		pcm := make([]int16, len(buf)/2)
		for i := range pcm {
			pcm[i] = int16(binary.LittleEndian.Uint16(buf[i*2:]))
		}
		opusFrame, err := cfg.Encoder.Encode(pcm, frameSize, len(pcm)*2)
		if err != nil {
			cfg.Logger.Println("Opus encode error:", err)
			continue
		}

		if cfg.VC != nil && cfg.VC.Ready {
			cfg.OpusSender <- opusFrame
		} else {
			cfg.Logger.Println("VC disconnected, stopping stream.")
			cfg.ControlChan <- controlMsg{Command: "stop"}
			return
		}

	}
}
