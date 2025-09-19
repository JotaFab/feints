package infra

import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"

	"feints/internal/core"
)

// DiscordPlayer une FeintsPlayer con Discord Voice
type DiscordPlayer struct {
	session   *discordgo.Session
	guildID   string
	channelID string
	vc        *discordgo.VoiceConnection
	player    *core.FeintsPlayer
	logger    *log.Logger
}

// NewDiscordPlayer inicializa un reproductor con Discord
func NewDiscordPlayer(s *discordgo.Session, guildID, channelID string, logger *log.Logger) (*DiscordPlayer, error) {
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return nil, err
	}

	dp := &DiscordPlayer{
		session:   s,
		guildID:   guildID,
		channelID: channelID,
		vc:        vc,
		player:    core.NewFeintsPlayer(logger),
		logger:    logger,
	}

	// loop que envía OutputCh → OpusSend
	go dp.streamLoop()

	return dp, nil
}

// streamLoop escucha el canal de salida del player y manda frames a Discord
func (dp *DiscordPlayer) streamLoop() {
	for {
		select {
		case frame, ok := <-dp.player.OutputCh:


			if !ok || dp.vc == nil || dp.vc.OpusSend == nil {
				continue // ignorar mientras no hay conexión
			}

			dp.vc.OpusSend <- frame // enviar frame
			// enviado correctamente
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// Métodos de control (puentes hacia CmdCh)
func (dp *DiscordPlayer) Add(song core.Song) {
	dp.player.CmdCh <- core.PlayerCommand{Name: "add", Arg: song}
}
func (dp *DiscordPlayer) Play() {
	dp.player.CmdCh <- core.PlayerCommand{Name: "play"}
}
func (dp *DiscordPlayer) Pause() {
	dp.player.CmdCh <- core.PlayerCommand{Name: "pause"}
}
func (dp *DiscordPlayer) Resume() {
	dp.player.CmdCh <- core.PlayerCommand{Name: "resume"}
}
func (dp *DiscordPlayer) Next() {
	dp.player.CmdCh <- core.PlayerCommand{Name: "next"}
}
func (dp *DiscordPlayer) Stop() {
	dp.player.CmdCh <- core.PlayerCommand{Name: "stop"}
}
func (dp *DiscordPlayer) QueueList() []core.Song {
	resp := make(chan any)
	dp.player.CmdCh <- core.PlayerCommand{Name: "list", Resp: resp}
	return (<-resp).([]core.Song)
}
