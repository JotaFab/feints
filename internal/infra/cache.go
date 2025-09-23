package infra

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"feints/internal/core"

	id3v2 "github.com/bogem/id3v2"
)

type SongCache struct {
	songs      map[string]*core.Song
	searches   map[string][]core.Song
	timestamps map[string]time.Time
	ttl        time.Duration
	muSongs    sync.RWMutex
	muSearch   sync.RWMutex
}

// --- PreloadSongCache ---
// Recorre SongsDir y carga las canciones en memoria si tienen metadatos ID3 válidos.
func PreloadSongCache(c *SongCache) error {
	entries, err := os.ReadDir(SongsDir)
	if err != nil {
		return fmt.Errorf("error leyendo directorio %s: %w", SongsDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".mp3") {
			continue
		}

		path := filepath.Join(SongsDir, entry.Name())
		tag, err := id3v2.Open(path, id3v2.Options{Parse: true})
		if err != nil {
			slog.Warn("ignorado archivo sin metadatos válidos",
				"file", entry.Name(), "error", err)
			continue
		}

		// Extraer metadata principal
		title := tag.Title()
		uploader := tag.Artist()
		length := tag.GetTextFrame(tag.CommonID("Length")).Text

		// Buscar frames TXXX (pueden contener URL u otros datos)
		var url string
		for _, f := range tag.GetFrames("TXXX") {
			if tf, ok := f.(id3v2.UserDefinedTextFrame); ok {
				slog.Info("frame TXXX encontrado", "description", tf.Description, "text", tf.Value)
				if strings.Contains(strings.ToLower(tf.Description), "url") || tf.Description == "" {
					url = tf.Value
				}
			}
		}

		_ = tag.Close()

		// Parsear duración si existe
		var dur time.Duration
		if n, _ := strconv.Atoi(length); n > 0 {
			dur = time.Duration(n) * time.Millisecond
		}

		// Crear Song y meter en cache
		s := core.Song{
			Title:    title,
			Uploader: uploader,
			Duration: dur,
			Path:     path,
			URL:      url,
		}
		c.AddSong(s)
	}
	return nil
}

var GlobalCache = NewSongCache(30 * time.Minute)

func NewSongCache(ttl time.Duration) *SongCache {
	return &SongCache{
		songs:      make(map[string]*core.Song),
		searches:   make(map[string][]core.Song),
		timestamps: make(map[string]time.Time),
		ttl:        ttl,
	}
}

func (c *SongCache) AddSong(s core.Song) {
	c.muSongs.Lock()
	defer c.muSongs.Unlock()
	if _, ok := c.songs[s.URL]; !ok {
		c.songs[s.URL] = &s
		log.Printf("[cache] Insertada canción: %s - %s", s.Uploader, s.Title)
	}
}

func (c *SongCache) GetSong(url string) *core.Song {
	c.muSongs.RLock()
	defer c.muSongs.RUnlock()
	return c.songs[url]
}

func (c *SongCache) AddSearch(query string, results []core.Song) {
	c.muSearch.Lock()
	defer c.muSearch.Unlock()
	c.searches[query] = results
	c.timestamps[query] = time.Now()
}

func (c *SongCache) GetSearch(query string) ([]core.Song, error) {
	c.muSearch.RLock()
	results, ok := c.searches[query]
	t, tsOk := c.timestamps[query]
	c.muSearch.RUnlock()
	if !ok || !tsOk || time.Since(t) > c.ttl {
		songs, err := Search(query, 5)
		if err == nil {
			c.AddSearch(query, songs)
			return c.GetSearch(query)
		}
		return nil, fmt.Errorf("somthing gone wrong in the GetSearch %v", err)
	}
	return results, nil
}
