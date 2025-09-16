// internal/player/player.go
package player

import (
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/bwmarrin/discordgo"
	"layeh.com/gopus"

	"feints/internal/queue"
)

// Player administra la reproducción de música en un canal de voz
type Player struct {
	sync.Mutex
	session   *discordgo.Session
	guildID   string
	channelID string
	vc        *discordgo.VoiceConnection
	queue     *queue.SongQueue

	encoder   *gopus.Encoder
	stopChan  chan struct{}
	pauseChan chan bool
	playing   bool
	current   *queue.Song
}

// NewPlayer crea un player ligado a un canal de voz
func NewPlayer(s *discordgo.Session, guildID, channelID string) (*Player, error) {
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return nil, fmt.Errorf("error al conectar al canal de voz: %w", err)
	}

	encoder, err := gopus.NewEncoder(48000, 2, gopus.Audio)
	if err != nil {
		return nil, fmt.Errorf("error creando encoder opus: %w", err)
	}

	return &Player{
		session:   s,
		guildID:   guildID,
		channelID: channelID,
		vc:        vc,
		queue:     queue.NewSongQueue(),
		encoder:   encoder,
		stopChan:  make(chan struct{}),
		pauseChan: make(chan bool, 1),
	}, nil
}

// PlaySong agrega una canción a la cola y si no hay reproducción activa, empieza el loop
func (p *Player) PlaySong(song queue.Song) {
	p.queue.Push(song)

	p.Lock()
	defer p.Unlock()

	if !p.playing {
		p.playing = true
		go p.loop()
	}
}

// loop procesa la cola continuamente
func (p *Player) loop() {
	for {
		next := p.queue.Pop()
		if next == nil {
			break
		}

		p.current = next
		if err := p.streamSong(next.URL); err != nil {
			fmt.Println("Error streaming:", err)
		}
	}

	p.Lock()
	p.playing = false
	p.current = nil
	p.Unlock()
}

// streamSong usa yt-dlp + ffmpeg y envía audio a Discord
func (p *Player) streamSong(query string) error {
	ytCmd := exec.Command("yt-dlp", "-f", "bestaudio", "-g", "ytsearch:"+query)
	out, err := ytCmd.Output()
	if err != nil {
		return fmt.Errorf("yt-dlp failed: %w", err)
	}
	audioURL := string(out)

	ffmpegCmd := exec.Command("ffmpeg",
		"-i", audioURL,
		"-f", "s16le",
		"-ar", "48000",
		"-ac", "2",
		"pipe:1",
	)
	ffmpegOut, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("ffmpeg stdout: %w", err)
	}
	if err := ffmpegCmd.Start(); err != nil {
		return fmt.Errorf("ffmpeg start: %w", err)
	}

	const frameSize = 960
	pcmBuf := make([]byte, frameSize*2*2)

	for {
		select {
		case <-p.stopChan:
			ffmpegCmd.Process.Kill()
			return nil
		default:
		}

		_, err := io.ReadFull(ffmpegOut, pcmBuf)
		if err != nil {
			break
		}

		// pause
		select {
		case paused := <-p.pauseChan:
			if paused {
				for {
					paused = <-p.pauseChan
					if !paused {
						break
					}
				}
			}
		default:
		}

		// encode
		pcm16 := make([]int16, len(pcmBuf)/2)
		for i := range pcm16 {
			pcm16[i] = int16(pcmBuf[i*2]) | int16(pcmBuf[i*2+1])<<8
		}
		opusFrame, err := p.encoder.Encode(pcm16, frameSize, len(pcm16)*2)
		if err != nil {
			return err
		}

		p.vc.OpusSend <- opusFrame
	}

	return ffmpegCmd.Wait()
}

// Pause detiene temporalmente la reproducción
func (p *Player) Pause() {
	select {
	case p.pauseChan <- true:
	default:
	}
}

// Resume reanuda la reproducción
func (p *Player) Resume() {
	select {
	case p.pauseChan <- false:
	default:
	}
}

// Stop cancela todo y limpia la cola
func (p *Player) Stop() {
	p.queue.Clear()
	close(p.stopChan)
	p.stopChan = make(chan struct{})
}

// Skip mata la canción actual y pasa a la siguiente
func (p *Player) Skip() {
	select {
	case p.stopChan <- struct{}{}:
	default:
	}
	p.stopChan = make(chan struct{})
}

// ShowQueue devuelve la cola actual
func (p *Player) ShowQueue() []queue.Song {
	return p.queue.List()
}

// NowPlaying devuelve la canción actual
func (p *Player) NowPlaying() *queue.Song {
	p.Lock()
	defer p.Unlock()
	return p.current
}
