package player

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"layeh.com/gopus"
)

// Song representa una canción
type Song struct {
	Title string
	URL   string
}

// DiscordPlayer administra la reproducción de música en un canal de voz
type DiscordPlayer struct {
	sync.Mutex
	session   *discordgo.Session
	guildID   string
	channelID string
	vc        *discordgo.VoiceConnection
	queue     *SongQueue
	encoder   *gopus.Encoder

	ctx       context.Context
	cancel    context.CancelFunc
	pauseChan chan bool
	playing   bool
	current   *Song
	status    string
}

// PlayerManager administra múltiples reproductores
type PlayerManager struct {
	Players map[string]*DiscordPlayer // clave = guildID+channelID
}

// NewPlayerManager crea un nuevo administrador
func NewPlayerManager() *PlayerManager {
	return &PlayerManager{
		Players: make(map[string]*DiscordPlayer),
	}
}

// GetPlayer obtiene un reproductor existente o crea uno nuevo si no existe
func (pm *PlayerManager) GetPlayer(s *discordgo.Session, guildID, channelID string) (*DiscordPlayer, error) {
	key := guildID + channelID

	// Verificar si ya existe
	if dp, ok := pm.Players[key]; ok {
		// Verificar si la conexión de voz sigue activa
		if dp.vc != nil && dp.vc.Ready {
			return dp, nil
		}
		// Si la VC no está activa, desconectar y eliminar para recrear
		_ = dp.Disconnect()
		delete(pm.Players, key)
	}

	// Intentar unirse al canal de voz
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return nil, fmt.Errorf("error al conectar al canal de voz: %w", err)
	}

	encoder, err := gopus.NewEncoder(48000, 2, gopus.Audio)
	if err != nil {
		return nil, fmt.Errorf("error creando encoder opus: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	dp := &DiscordPlayer{
		session:   s,
		guildID:   guildID,
		channelID: channelID,
		vc:        vc,
		queue:     NewSongQueue(),
		encoder:   encoder,
		ctx:       ctx,
		cancel:    cancel,
		pauseChan: make(chan bool, 1),
		status:    "stopped",
	}

	pm.Players[key] = dp
	return dp, nil
}


// Disconnect cierra la conexión de voz
func (p *DiscordPlayer) Disconnect() error {
	p.cancel()
	if p.vc != nil {
		return p.vc.Disconnect()
	}
	return nil
}

// PlaySong agrega una canción a la cola y lanza el loop si no está corriendo
func (p *DiscordPlayer) PlaySong(song Song) error {
	p.queue.Push(song)

	p.Lock()
	defer p.Unlock()
	if !p.playing {
		p.playing = true
		p.status = "playing"
		go p.loop()
	}
	return nil
}

// loop procesa la cola continuamente
func (p *DiscordPlayer) loop() {
	for {
		select {
		case <-p.ctx.Done():
			return
		default:
		}

		next := p.queue.Pop()
		if next == nil {
			break
		}

		p.Lock()
		p.current = next
		p.Unlock()

		if err := p.streamSong(next.URL); err != nil {
			fmt.Println("Error streaming:", err)
		}
	}

	p.Lock()
	p.playing = false
	p.current = nil
	p.status = "stopped"
	p.Unlock()
}

// streamSong usa yt-dlp + ffmpeg y envía audio a Discord
func (p *DiscordPlayer) streamSong(query string) error {
	ytCmd := exec.Command("yt-dlp", "-f", "bestaudio", "-g", "ytsearch:"+query)
	out, err := ytCmd.Output()
	if err != nil {
		return fmt.Errorf("yt-dlp failed: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 {
		return fmt.Errorf("no se encontró audio")
	}
	audioURL := lines[0]

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

	_ = p.vc.Speaking(true)
	defer p.vc.Speaking(false)

	for {
		select {
		case <-p.ctx.Done():
			ffmpegCmd.Process.Kill()
			return nil
		default:
		}

		_, err := io.ReadFull(ffmpegOut, pcmBuf)
		if err != nil {
			break
		}

		// pause handling
		select {
		case paused := <-p.pauseChan:
			if paused {
				p.Lock()
				p.status = "paused"
				p.Unlock()
				// esperar a reanudar
				for {
					paused = <-p.pauseChan
					if !paused {
						p.Lock()
						p.status = "playing"
						p.Unlock()
						break
					}
				}
			}
		default:
		}

		// encode
		pcm16 := make([]int16, len(pcmBuf)/2)
		for i := range pcm16 {
			pcm16[i] = int16(binary.LittleEndian.Uint16(pcmBuf[i*2:]))
		}
		opusFrame, err := p.encoder.Encode(pcm16, frameSize, len(pcm16)*2)
		if err != nil {
			return err
		}

		p.vc.OpusSend <- opusFrame
	}

	return ffmpegCmd.Wait()
}

// Pause pausa
func (p *DiscordPlayer) Pause() error {
	select {
	case p.pauseChan <- true:
	default:
	}
	return nil
}

// Resume reanuda
func (p *DiscordPlayer) Resume() error {
	select {
	case p.pauseChan <- false:
	default:
	}
	return nil
}

// Skip salta la canción actual pero mantiene el loop
func (p *DiscordPlayer) Skip() error {
	p.cancel()                          // cancela canción actual
	p.ctx, p.cancel = context.WithCancel(context.Background()) // crea nuevo contexto para el loop
	return nil
}

// Current devuelve la canción actual
func (p *DiscordPlayer) Current() *Song {
	p.Lock()
	defer p.Unlock()
	return p.current
}

// Queue devuelve la cola
func (p *DiscordPlayer) Queue() []Song {
	return p.queue.List()
}

// Clear limpia la cola
func (p *DiscordPlayer) Clear() error {
	p.queue.Clear()
	return nil
}

// IsPlaying retorna true si está activo
func (p *DiscordPlayer) IsPlaying() bool {
	p.Lock()
	defer p.Unlock()
	return p.playing
}

// Status retorna el estado actual
func (p *DiscordPlayer) Status() string {
	p.Lock()
	defer p.Unlock()
	return p.status
}
