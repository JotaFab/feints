package player

import (
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// --- Tipos y constantes ---
type PlayerState string

const (
	Idle    PlayerState = "idle"
	Playing PlayerState = "playing"
	Paused  PlayerState = "paused"
)

// --- Player interface ---
type Player interface {
	AddToQueue(song Song)
	Play()
	Pause()
	Stop()
	NextSong()
	NowPlaying() Song
	QueueList() []Song
	Clear()
	Status() string
}

// --- Comandos para el canal de control ---
type controlMsg struct {
	Command string
}

// --- internal DiscordPlayer ---
type discordPlayer struct {
	sync.RWMutex

	session    *discordgo.Session
	guildID    string
	channelID  string
	vc         *discordgo.VoiceConnection

	queue   *SongQueue
	state   PlayerState
	current *Song
	logger  *log.Logger

	control    chan controlMsg
	stopStream chan struct{}
}

// --- Implementación de la interfaz Player ---
func (dp *discordPlayer) AddToQueue(song Song) {
    dp.queue.Push(song)
}

func (dp *discordPlayer) Play() {
    dp.control <- controlMsg{Command: "play"}
}

func (dp *discordPlayer) Pause() {
    dp.control <- controlMsg{Command: "pause"}
}

func (dp *discordPlayer) Stop() {
    dp.control <- controlMsg{Command: "stop"}
}

func (dp *discordPlayer) NextSong() {
    dp.control <- controlMsg{Command: "skip"}
}

func (dp *discordPlayer) NowPlaying() Song {
    dp.RLock()
    defer dp.RUnlock()
    if dp.current != nil {
        return *dp.current
    }
    return Song{}
}

func (dp *discordPlayer) QueueList() []Song {
    dp.RLock()
    defer dp.RUnlock()
    list := dp.queue.List()
    return list
}

func (dp *discordPlayer) Status() string {
    dp.RLock()
    defer dp.RUnlock()
    return string(dp.state)
}

func (dp *discordPlayer) startStream(song *Song) {
	
	
	
	dp.Lock()
	dp.stopCurrentStream()
	disconnectVoice(dp.vc)
		var err error
		dp.vc, err = connectVoice(dp.session, dp.guildID, dp.channelID)
		if err != nil {
			dp.logger.Println("Error connecting to voice:", err)
			return
		}
	dp.current = song
	dp.state = Playing
	dp.Unlock()
	dp.logger.Println("se esta estremeando una nueva cancion?")

	// Configura el struct para pasar la información necesaria
	cfg := &StreamConfig{
		VC:          dp.vc,
		OpusSender:  dp.vc.OpusSend,
		CurrentSong: dp.current,
		StopChan:    dp.stopStream,
		ControlChan: dp.control,
		State:       dp.state,
		Logger:      dp.logger,
		StateMu:     &dp.RWMutex,
	}

	go StreamSongLoop(cfg)
}

func (dp *discordPlayer) stopCurrentStream() {
	select {
	case dp.stopStream <- struct{}{}:
	default:
	}
}

// --- Constructor ---
func NewDiscordPlayer(s *discordgo.Session, guildID, channelID string) Player {
	

	dp := &discordPlayer{
		session:    s,
		guildID:    guildID,
		channelID:  channelID,
		queue:      NewSongQueue(),
		state:      Idle,
		control:    make(chan controlMsg, 10),
		stopStream: make(chan struct{}, 1),
		logger:     log.New(log.Writer(), "["+guildID+"] ", log.LstdFlags),
		current:    nil,
	}
	go dp.stateLoopRun()
	return dp
}

// --- Loop de control (el actor principal) ---
func (dp *discordPlayer) stateLoopRun() {
	for msg := range dp.control {
		dp.logger.Println(msg)
		dp.handleControl(msg)
	}
}

func (dp *discordPlayer) handleControl(msg controlMsg) {
	dp.Lock()
	defer dp.Unlock()

	switch msg.Command {
	case "play":
		if dp.state == Idle {
			dp.logger.Println("Starting playback...")
			if next := dp.queue.Pop(); next != nil {
				// Aquí se inicia la reproducción, el actor principal es el encargado
				dp.current = next
				dp.state = Playing
				go dp.startStream(next)
			}
		}
	case "pause":
		if dp.state == Playing {
			dp.state = Paused
			dp.logger.Println("Paused.")
		}
	case "resume":
		if dp.state == Paused {
			dp.state = Playing
			dp.logger.Println("Resumed.")
		}
	case "stop":
		select {
		case dp.stopStream <- struct{}{}:
		default:
		}
		dp.Clear()
		dp.logger.Println("Stopped.")
	case "skip":
		select {
		case dp.stopStream <- struct{}{}:
		default:
		}
		if next := dp.queue.Pop(); next != nil {
			dp.current = next
			dp.state = Playing
		} else {
			dp.state = Idle
			dp.current = nil
		}
		dp.logger.Println("Skipped.")
	case "end_of_stream":
		if next := dp.queue.Pop(); next != nil {
			dp.current = next
			dp.state = Playing
			go dp.stopCurrentStream()
		} else {
			dp.state = Idle
			dp.current = nil
			select {
			case dp.stopStream <- struct{}{}:
			default:
			}
		}
	}
}

func (dp *discordPlayer) Clear() {
    dp.Lock()
    defer dp.Unlock()
	go dp.stopCurrentStream()


    dp.queue.Clear() // Call the Clear method on the SongQueue
    dp.current = nil // Reset the current song
    dp.state = Idle  // Set the player state to idle
}