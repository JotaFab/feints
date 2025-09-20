package infra

import (
	"log"
	"sync"
	"time"

	"feints/internal/core"
)

type SongCache struct {
	songs      map[string]*core.Song
	searches   map[string][]core.Song
	timestamps map[string]time.Time
	ttl        time.Duration
	muSongs    sync.RWMutex
	muSearch   sync.RWMutex
}

func NewSongCache(ttl time.Duration) *SongCache {
	return &SongCache{
		songs:      make(map[string]*core.Song),
		searches:   make(map[string][]core.Song),
		timestamps: make(map[string]time.Time),
		ttl:        ttl,
	}
}

func (c *SongCache) AddSong(s *core.Song) {
	c.muSongs.Lock()
	defer c.muSongs.Unlock()
	if _, ok := c.songs[s.Path]; !ok {
		c.songs[s.Path] = s
		log.Printf("[cache] Insertada canción: %s - %s", s.Uploader, s.Title)
	}
}

func (c *SongCache) GetSong(path string) *core.Song {
	c.muSongs.RLock()
	defer c.muSongs.RUnlock()
	return c.songs[path]
}

func (c *SongCache) AddSearch(query string, results []core.Song) {
	c.muSearch.Lock()
	defer c.muSearch.Unlock()
	c.searches[query] = results
	c.timestamps[query] = time.Now()
}

func (c *SongCache) GetSearch(query string) ([]core.Song, bool) {
	c.muSearch.RLock()
	defer c.muSearch.RUnlock()
	results, ok := c.searches[query]
	t, tsOk := c.timestamps[query]
	if !ok || !tsOk || time.Since(t) > c.ttl {
		return nil, false
	}
	return results, true
}
