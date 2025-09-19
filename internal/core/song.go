package core

import "time"

// Song representa una canción genérica dentro del dominio.
type Song struct {
	Title     string
	Uploader  string
	Album     string
	Duration  time.Duration
	Thumbnail string
	Path      string // Ruta local del archivo de audio (mp3, wav, etc.)
	URL       string // URL original del video o fuente
}