package core

import (
	"log"
	"os/exec"
)

// PlayerState define el estado actual del reproductor
type PlayerState string

const (
	StateIdle    PlayerState = "idle"
	StatePlaying PlayerState = "playing"
	StatePaused  PlayerState = "paused"
)

// PlayerCommand representa un comando de control
type PlayerCommand struct {
	Name string
	Arg  any
	Resp chan any
}

// FeintsPlayer maneja la cola, reproducción y comandos
type FeintsPlayer struct {
	Queue     *SongQueue
	CmdCh     chan PlayerCommand
	OutputCh  chan []byte
	state     PlayerState
	current   *Song
	logger    *log.Logger
	cancelCmd *exec.Cmd
	skipCh    chan struct{}
}

// NewFeintsPlayer crea un reproductor nuevo
func NewFeintsPlayer(logger *log.Logger) *FeintsPlayer {
	p := &FeintsPlayer{
		Queue:    NewSongQueue(),
		CmdCh:    make(chan PlayerCommand),
		OutputCh: make(chan []byte, 10),
		state:    StateIdle,
		logger:   logger,
		skipCh:   make(chan struct{}, 1),
	}
	go p.loop()
	return p
}

// loop principal: maneja todos los comandos
func (p *FeintsPlayer) loop() {
	for cmd := range p.CmdCh {
		switch cmd.Name {
		case "add":
			song := cmd.Arg.(Song)
			s, err := YtdlpBestAudioURL(song.URL)
			if err != nil {
				p.logger.Println(err)
				continue
			}
			p.Queue.Push(*s)
			p.logger.Printf("Song added: %s", s.Title)

			// si está idle, comenzar a reproducir
			if p.state == StateIdle {
				p.current = p.Queue.Pop()
				if p.current != nil {
					p.state = StatePlaying
					go p.streamSong(p.current)
				}
			}

		case "pause":
			if p.state == StatePlaying {
				p.state = StatePaused
				p.logger.Println("Paused playback")
			}

		case "resume":
			if p.state == StatePaused {
				p.state = StatePlaying
				p.logger.Println("Resumed playback")
			}

		case "next", "skip":

			if p.cancelCmd != nil {
				_ = p.cancelCmd.Process.Kill()
				select {
				case p.skipCh <- struct{}{}:
				default:
				}
			}

		case "stop":
			if p.cancelCmd != nil {
				_ = p.cancelCmd.Process.Kill()
			}
			p.state = StateIdle
			p.current = nil
			p.Queue.Clear()
			p.logger.Println("Stopped and cleared queue")

		case "list":
			p.logger.Println(p.Queue.List())
			cmd.Resp <- p.Queue.List()
		}
	}
}


func (p *FeintsPlayer) streamSong(song *Song) {
	p.logger.Printf("Now streaming: %s", song.Title)

	opusCh := make(chan []byte, 10)
	cmd, err := StreamFromPathToOpusChan(song.Path, opusCh)
	if err != nil {
		p.logger.Printf("Error starting stream: %v", err)
		return
	}
	p.cancelCmd = cmd

	for {
		select {
		case frame, ok := <-opusCh:
			if !ok || p.state != StatePlaying {
				p.cancelCmd = nil
				return
			}
			p.OutputCh <- frame
		case <-p.skipCh:
			p.logger.Printf("Skipped: %s", song.Title)
			p.cancelCmd.Process.Kill()
			p.cancelCmd = nil
			return
		}
	}
	
}

