// internal/audio/audio.go
package audio

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"

	"layeh.com/gopus"

	"feints/internal/context"
)

type SongMetadata struct {
	Title       string `json:"title"`
	Uploader    string `json:"uploader"`
	Duration    int    `json:"duration"` // en segundos
	WebpageURL  string `json:"webpage_url"`
	Thumbnail   string `json:"thumbnail"`
}

func GetSongMetadata(query string) (*SongMetadata, error) {
	// Usar yt-dlp para obtener metadata en JSON
	cmd := exec.Command("yt-dlp", "-j", "ytsearch:"+query)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	var meta SongMetadata
	if err := json.Unmarshal(out, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}
	return &meta, nil
}

// GetAudioStream obtiene el audio de YouTube y devuelve un io.ReadCloser de PCM
func GetAudioStream(song string) (io.ReadCloser, *exec.Cmd, error) {
	ytCmd := exec.Command("yt-dlp", "-f", "bestaudio", "-g", "ytsearch:"+song)
	out, err := ytCmd.Output()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to run yt-dlp: %w", err)
	}
	audioURL := string(out)
	audioURL = string([]byte(audioURL)) // limpiar cualquier whitespace

	ffmpegCmd := exec.Command("ffmpeg",
		"-i", audioURL,
		"-f", "s16le",
		"-ar", "48000",
		"-ac", "2",
		"pipe:1",
	)
	ffmpegOut, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get ffmpeg stdout: %w", err)
	}
	if err := ffmpegCmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("failed to start ffmpeg: %w", err)
	}
	return ffmpegOut, ffmpegCmd, nil
}

// SendAudioStream lee PCM de ffmpeg y lo envía a Discord
func SendAudioStream(ctx *context.VoiceContext, ffmpegOut io.ReadCloser, ffmpegCmd *exec.Cmd) error {
	const (
		sampleRate = 48000
		channels   = 2
		frameSize  = 960
	)

	encoder, err := gopus.NewEncoder(sampleRate, channels, gopus.Audio)
	if err != nil {
		return fmt.Errorf("failed to create opus encoder: %w", err)
	}

	ctx.Mutex.Lock()
	ctx.Playing = true
	if ctx.StopChan == nil {
		ctx.StopChan = make(chan struct{})
	}
	ctx.Mutex.Unlock()

	pcmBuf := make([]byte, frameSize*channels*2)
	for {
		_, err := io.ReadFull(ffmpegOut, pcmBuf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read PCM: %w", err)
		}

		pcm16 := make([]int16, frameSize*channels)
		for i := 0; i < len(pcm16); i++ {
			pcm16[i] = int16(binary.LittleEndian.Uint16(pcmBuf[i*2:]))
		}

		opusFrame, err := encoder.Encode(pcm16, frameSize, len(pcm16)*2)
		if err != nil {
			return fmt.Errorf("failed to encode opus: %w", err)
		}

		select {
		case ctx.VC.OpusSend <- opusFrame:
		case <-ctx.StopChan:
			ffmpegCmd.Process.Kill()
			ctx.Mutex.Lock()
			ctx.Playing = false
			ctx.Mutex.Unlock()
			return nil
		}
	}

	ctx.Mutex.Lock()
	ctx.Playing = false
	ctx.Mutex.Unlock()

	if err := ffmpegCmd.Wait(); err != nil && err.Error() != "signal: killed" {
		return fmt.Errorf("ffmpeg exited with error: %w", err)
	}

	return nil
}
