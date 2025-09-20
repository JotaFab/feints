package infra

import (
	"feints/internal/core"
	"log/slog"
	"time"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
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
	queueA    []*core.Song
	Queue     chan core.Song  `json:"-"`
	Control   chan controlCmd `json:"-"`
	state     PlayerState
	Logger    *slog.Logger `json:"-"`
	vc        *discordgo.VoiceConnection
	stopCh    chan bool
	doneCh    chan bool
	autoplay  bool
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
		stopCh:    make(chan bool),
		doneCh:    make(chan bool),
	}
	go p.stateLoop()
	return p
}

// --- Interface methods ---
func (p *DgvoicePlayer) Play()     { p.Control <- cmdPlay }
func (p *DgvoicePlayer) Next()     { p.Control <- cmdNext }
func (p *DgvoicePlayer) Pause()    { p.Control <- cmdPause }
func (p *DgvoicePlayer) Resume()   { p.Control <- cmdResume }
func (p *DgvoicePlayer) Stop()     { p.Control <- cmdStop }
func (p *DgvoicePlayer) AutoPlay() { p.autoplay = true }

func (p *DgvoicePlayer) AddSong(song core.Song) {
	p.Logger.Info("Queueing song", "title", song.Title)
	p.Queue <- song
	p.queueA = append(p.queueA, &song)
}

func (p *DgvoicePlayer) ListQueue() []*core.Song {
	snapshot := make([]*core.Song, len(p.queueA))
	copy(snapshot, p.queueA)
	return snapshot
}

func (p *DgvoicePlayer) State() string { return string(p.state) }

// --- Bucle central ---
func (p *DgvoicePlayer) stateLoop() {
	p.Logger.Info("State loop started")

	for {
		select {
		case cmd := <-p.Control:
			p.cmdHandler(cmd)
		default:
		}
		switch p.state {
		case Idle:
			select {
			case song := <-p.Queue:
				p.Logger.Info("Auto-playing next song", "title", song.Title)
				p.state = Playing
				go p.playSong(song)
			default:
				if p.autoplay {
					s, e := GlobalSongService.GetRandomLocalSong()
					if e != nil {
						p.Logger.Error("error getting random local song", "error", e)
						p.Stop()
						continue
					}
					p.AddSong(*s)
				}
			}

		case Playing:
			select {
			case <-p.doneCh:
				p.Logger.Info("Song finished")
				p.state = Idle
			default:
			}
		case Paused:
			time.Sleep(time.Second)
		case Stopped:
			time.Sleep(time.Second)

		default:
		}

	}
}

// --- Manejo de comandos ---
func (p *DgvoicePlayer) cmdHandler(cmd controlCmd) {
	switch cmd {
	case cmdPlay:
		if p.state != Playing {
			select {
			case song := <-p.Queue:
				p.Logger.Info("Auto-playing next song", "title", song.Title)
				p.state = Playing
				go p.playSong(song)
			default:
			}

		}

	case cmdPause:
		p.state = Paused
		p.Logger.Info("Paused")
		if p.vc != nil {
			p.vc.Speaking(false)
		}

	case cmdResume:
		p.state = Playing
		p.Logger.Info("Resumed")
		if p.vc != nil {
			p.vc.Speaking(true)
		}

	case cmdNext:
		p.Logger.Info("Skipping song")
		p.stopCurrentPlayback()
		p.state = Idle

	case cmdStop:
		p.Logger.Info("Stopping and clearing queue")
		p.stopCurrentPlayback()
		p.Queue = make(chan core.Song, 50)
		p.queueA = make([]*core.Song,0)
		p.autoplay = false
		p.state = Idle
	}
}
func (p *DgvoicePlayer) stopCurrentPlayback() {
	// Use a non-blocking select to send stop signal
	select {
	case p.stopCh <- true:
	default:
	}
	if p.vc != nil {
		p.vc.Speaking(false)
		p.vc.Disconnect()
		p.vc = nil
	}
}

// --- Reproducir canción ---
func (p *DgvoicePlayer) playSong(song core.Song) {
	var err error

	if len(p.queueA) > 0 {
		p.queueA = p.queueA[1:]
	} else {
		p.Logger.Debug("no more songs in queue")
	}
	s:= GlobalSongService.cache.GetSong(song.URL)
	if s != nil {
		song = *s
	} else {

		s, err := GlobalSongService.SongReadyToPlay(song)
		
		if err != nil {
			p.Logger.Error("error downloading the song", "error", err)
			p.state = Stopped
			return
		}
		song = *s
	}

	p.Logger.Info("Playing song", "title", song.Title)
	p.stopCurrentPlayback() // Retry joining voice channel with a small delay
	var vc *discordgo.VoiceConnection
	for i := 0; i < 3; i++ { // Try up to 3 times
		vc, err = p.Session.ChannelVoiceJoin(p.GuildID, p.ChannelID, false, true)
		if err == nil {
			break // Success
		}
		p.Logger.Error("Join failed, retrying...", "error", err, "attempt", i+1)
		time.Sleep(2 * time.Second) // Wait before retrying
	}
	if err != nil {
		p.Logger.Error("Join failed", "error", err)
		p.state = Idle
		return
	}
	p.vc = vc
	p.vc.Speaking(true)

	dgvoice.PlayAudioFile(p.vc, song.Path, p.stopCh)
	// al terminar la canción
	p.doneCh <- true
}
