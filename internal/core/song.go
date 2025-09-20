package core

import "time"

// Song representa una canción genérica dentro del dominio.
type Song struct {
	Title     string        `json:"title"`
	Uploader  string        `json:"uploader"`
	Duration  time.Duration `json:"duration"`
	Thumbnail string        `json:"thumbnail"`
	URL       string        `json:"url"`
	Path      string        `json:"path"`
}