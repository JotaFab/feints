package infra

import (
	"feints/internal/core"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	id3v2 "github.com/bogem/id3v2"
)

const SongsDir = "songs"

type SongService struct {
	cache *SongCache
}

func NewSongService(c *SongCache) *SongService {
	PreloadSongCache(c)
	return &SongService{cache: c}
}

var GlobalSongService = NewSongService(GlobalCache)

func (s *SongService) SongReadyToPlay(song core.Song) (*core.Song, error) {
	

	// 1. Si viene con URL, obtener metadata
	if song.URL == "" {
		return nil, fmt.Errorf("song no tiene URL ni Path")
	}

	meta, err := Metadata(song.URL)
	if err != nil {
		return nil, err
	}

	filename := sanitizeFilename(fmt.Sprintf("%s-%s.mp3", meta.Uploader, meta.Title))
	path := filepath.Join(SongsDir, filename)
	if err := DownloadAudio(meta.URL, path); err != nil {
		return nil, err
	}
	meta.Path = path

	return meta, nil
}

// sanitize helper
func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, ":", "_")
	return name
}

// GetRandomLocalSong devuelve una canción aleatoria del directorio local de canciones.
func (s *SongService) GetRandomLocalSong() (*core.Song, error) {
	files, err := os.ReadDir(SongsDir)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer SongsDir: %w", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no hay canciones en SongsDir")
	}

	// Escoger aleatoriamente un archivo
	f := files[rand.Intn(len(files))]
	path := filepath.Join(SongsDir, f.Name())

	tag, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		return nil, fmt.Errorf("error leyendo metadata de %s: %w", path, err)
	}
	defer tag.Close()

	length := tag.GetTextFrame(tag.CommonID("Length")).Text
	var dur time.Duration
	if n, _ := strconv.Atoi(length); n > 0 {
		dur = time.Duration(n) * time.Second
	}
	// Buscar frames TXXX (pueden contener URL u otros datos)
		var url string
		for _, f := range tag.GetFrames("TXXX") {
			if tf, ok := f.(id3v2.UserDefinedTextFrame); ok {
				if strings.Contains(strings.ToLower(tf.Description), "url") || tf.Description == "" {
					url = tf.Value
				}
			}
		}
	song := &core.Song{
		Title:    tag.Title(),
		Uploader: tag.Artist(),
		Duration: dur,
		Path:     path,
		URL:      url,
	}
	return song, nil
}
