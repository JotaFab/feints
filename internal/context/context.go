package context

import (
	"os/exec"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type VoiceContext struct {
	GuildID   string
	ChannelID string
	VC        *discordgo.VoiceConnection
	Queue     *SongQueue
	AudioChan chan []byte   // canal para enviar PCM a la goroutine de audio
	FFmpegCmd *exec.Cmd     // proceso ffmpeg actual
	Mutex     sync.Mutex
	StopChan   chan struct{} 
	Playing	bool
}

var (
	contexts   = make(map[string]*VoiceContext)
	contextsMu sync.Mutex
)

// GetOrCreateContext crea o devuelve un contexto de voz existente
func GetOrCreateContext(s *discordgo.Session, guildID, channelID string) (*VoiceContext, error) {
	contextsMu.Lock()
	defer contextsMu.Unlock()

	if ctx, ok := contexts[guildID]; ok {
		return ctx, nil
	}

	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return nil, err
	}

	ctx := &VoiceContext{
		GuildID:   guildID,
		ChannelID: channelID,
		VC:        vc,
		Queue:     NewSongQueue(),
		Playing: false,
		AudioChan: make(chan []byte, 10), // canal bufferizado
	}
	contexts[guildID] = ctx
	return ctx, nil
}

// RemoveContext elimina un contexto de voz
func RemoveContext(guildID string) {
	contextsMu.Lock()
	defer contextsMu.Unlock()
	delete(contexts, guildID)
}
