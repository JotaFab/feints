package player


type PlayerManager struct {
	Players map[string]Player
}

func NewPlayerManager() *PlayerManager {
	return &PlayerManager{Players: make(map[string]Player, 0)}
}


// Song representa una canción en la cola.
type Song struct {
	Title string
	URL   string
	index int // Usado por container/heap
}
