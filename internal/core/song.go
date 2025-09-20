package core

import (
	"fmt"
	"time"
)

// Song representa una canción genérica dentro del dominio.
type Song struct {
	Title     string        `json:"title"`
	Uploader  string        `json:"uploader"`
	Duration  time.Duration `json:"duration"`
	Thumbnail string        `json:"thumbnail"`
	URL       string        `json:"url"`
	Path      string        `json:"path"`
}

// String devuelve una representación legible de la canción.
func (s Song) String() string {
	return fmt.Sprintf(
		"Song{Title=%q, Uploader=%q, Duration=%s, URL=%s, Path=%s}",
		s.Title, s.Uploader, s.Duration.String(), s.URL, s.Path,
	)
}