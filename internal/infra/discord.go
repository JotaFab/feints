package infra

import (
	"feints/internal/core"
	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"log/slog"
)

// --- PlayerState ---
type PlayerState string

const (
	Idle    PlayerState = "idle"
	Playing PlayerState = "playing"
	Paused  PlayerState = "paused"
	Stopped PlayerState = "stopped"
)

type DgvoicePlayer struct {
	Session   *discordgo.Session `json:"-"`
	GuildID   string             `json:"guild_id"`
	ChannelID string             `json:"channel_id"`
	Queue     chan core.Song     `json:"-"`
	Control   chan controlCmd    `json:"-"`
	state     PlayerState
	Logger    *slog.Logger `json:"-"`
	vc        *discordgo.VoiceConnection
	stopCh    chan bool
}

type controlCmd string

const (
	cmdPlay   controlCmd = "play"
	cmdPause  controlCmd = "pause"
	cmdResume controlCmd = "resume"
	cmdNext   controlCmd = "next"
	cmdStop   controlCmd = "stop"
)

// NewDgvoicePlayer devuelve un Player
func NewDgvoicePlayer(session *discordgo.Session, guildID, channelID string, l *slog.Logger) core.Player {
	p := &DgvoicePlayer{
		Session:   session,
		GuildID:   guildID,
		ChannelID: channelID,
		Queue:     make(chan core.Song, 50),
		Control:   make(chan controlCmd),
		state:     Idle,
		Logger:    l.With("component", "Player", "guild", guildID),
		stopCh:    make(chan bool, 1),
	}
	go p.loop()
	return p
}

// --- Interface methods ---
func (p *DgvoicePlayer) Play() {
	p.Logger.Info("Play() called")
	p.Control <- cmdPlay
}

func (p *DgvoicePlayer) AddSong(song core.Song) {
	p.Logger.Info("Queueing song", "title", song.Title)
	p.Queue <- song
}

func (p *DgvoicePlayer) Next()   { p.Control <- cmdNext }
func (p *DgvoicePlayer) Pause()  { p.Control <- cmdPause }
func (p *DgvoicePlayer) Resume() { p.Control <- cmdResume }
func (p *DgvoicePlayer) Stop()   { p.Control <- cmdStop }

func (p *DgvoicePlayer) ListQueue() []core.Song {
	snapshot := []core.Song{}
	for {
		select {
		case s := <-p.Queue:
			snapshot = append(snapshot, s)
		default:
			return snapshot
		}
	}
}

func (p *DgvoicePlayer) State() string { return string(p.state) }

// --- loop privado ---
func (p *DgvoicePlayer) loop() {
	for cmd := range p.Control {
		switch cmd {
		case cmdPlay:
			song := <-p.Queue
			go p.playSong(song)
		case cmdPause:
			p.pause()
		case cmdResume:
			p.resume()
		case cmdNext:
			p.Logger.Info("Skipping song")
			if p.vc != nil {
				p.vc.Speaking(false)
			}
			p.stopCh <- true

		case cmdStop:
			p.Logger.Info("Stopping and clearing queue")
			if p.vc != nil {
				p.vc.Disconnect()
			}
			p.stopCh <- true
			close(p.Queue)
			return
		}
	}
}

func (p *DgvoicePlayer) playSong(song core.Song) {

	s, err := YtdlpBestAudioURL(song.URL)
	if err != nil {
		p.Logger.Error("error downloading the song", "error", err)
		return
	}

	p.state = Playing
	p.Logger.Info("Playing song", "title", s.Title)
	if p.vc == nil {
		vc, err := p.Session.ChannelVoiceJoin(p.GuildID, p.ChannelID, false, true)
		if err != nil {
			p.Logger.Error("Join failed", "error", err)
			return
		}
		p.vc = vc
	}
	dgvoice.PlayAudioFile(p.vc, s.Path, p.stopCh)
	p.Logger.Info("Finished", "title", s.Title)
	p.state = Idle
}

func (p *DgvoicePlayer) pause() {
	if p.state == Playing {
		p.state = Paused
		p.Logger.Info("Paused")
		if p.vc != nil {
			p.vc.Speaking(false)
		}
	}
}

func (p *DgvoicePlayer) resume() {
	if p.state == Paused {
		p.state = Playing
		p.Logger.Info("Resumed")
		if p.vc != nil {
			p.vc.Speaking(true)
		}
	}
}
