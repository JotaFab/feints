package core


// --- Player interface ---
// Esto es lo que vas a usar en tu bot sin importar la implementación
type Player interface {
	Play()
	AddSong(song Song)
	Next()
	Pause()
	Resume()
	Stop()
	ListQueue() []Song
	State() string
	AutoPlay()
}
