package infra

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	id3v2 "github.com/bogem/id3v2"

	_ "image/jpeg"
	_ "image/png"
)



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
type ytDlpFlatItem struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Webpage string `json:"webpage_url"`
}

// -------------------- Variables globales --------------------

var (
	searchCache   = make(map[string][]Song)
	searchCacheMu sync.RWMutex
	searchLock    sync.Mutex
	searchTTL     = 10 * time.Minute
	cacheTime     = make(map[string]time.Time)

	songCache   []Song
	songCacheMu sync.RWMutex
)

const (
	YtDlpBin = "yt-dlp"
)

// -------------------- Funciones --------------------

// YtdlpSearch busca canciones en YouTube y guarda resultados en cache
func YtdlpSearch(query string, limit int) ([]Song, error) {
	if limit <= 0 || limit > 5 {
		limit = 5
	}

	// Revisa cache primero
	searchCacheMu.RLock()
	cached, ok := searchCache[query]
	timestamp, tsOk := cacheTime[query]
	searchCacheMu.RUnlock()
	if ok && tsOk && time.Since(timestamp) < searchTTL {
		log.Printf("[yt-dlp] Cache hit for query: %s", query)
		return cached, nil
	}

	// Solo una búsqueda a la vez
	searchLock.Lock()
	defer searchLock.Unlock()

	// Rechequear cache después de lock
	searchCacheMu.RLock()
	cached, ok = searchCache[query]
	timestamp, tsOk = cacheTime[query]
	searchCacheMu.RUnlock()
	if ok && tsOk && time.Since(timestamp) < searchTTL {
		return cached, nil
	}

	// Ejecutar yt-dlp
	out, stderr, err := runYtDlp(
		"--dump-json",
		"--flat-playlist",
		fmt.Sprintf("ytsearch%d:%s", limit, query),
	)
	if err != nil {
		return nil, fmt.Errorf("yt-dlp search error: %w - %s", err, stderr)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	var results []Song
	for _, line := range lines {
		var item ytDlpFlatItem
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			continue
		}

		song := Song{
			Title: item.Title,
			URL:   item.Webpage,
		}
		if song.URL == "" {
			song.URL = "https://www.youtube.com/watch?v=" + item.ID
		}

		results = append(results, song)
		addSongToCache(song)
	}

	// Guardar en cache de búsqueda
	searchCacheMu.Lock()
	searchCache[query] = results
	cacheTime[query] = time.Now()
	searchCacheMu.Unlock()

	return results, nil
}

// addSongToCache evita duplicados
func addSongToCache(s Song) {
	songCacheMu.Lock()
	defer songCacheMu.Unlock()
	for _, cached := range songCache {
		if cached.URL == s.URL {
			return
		}
	}
	songCache = append(songCache, s)
}

// GetCachedSong busca una canción en memoria
func GetCachedSong(url string) *Song {
	songCacheMu.RLock()
	defer songCacheMu.RUnlock()
	for i := range songCache {
		if songCache[i].URL == url {
			return &songCache[i]
		}
	}
	return nil
}

// YtdlpBestAudioURL descarga y devuelve la ruta del MP3 + cache
func YtdlpBestAudioURL(videoURL string) (*Song, error) {


	out, _, err := runYtDlp(
		"--cookies", "cookies.txt",
		"--flat-playlist",
		"--print", "%(title)s|%(uploader)s|%(duration)s|%(thumbnail)s",
		videoURL,
	)
	if err != nil {
		return nil, fmt.Errorf("error metadata: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(out), "|")
	if len(parts) < 4 {
		return nil, fmt.Errorf("metadata incompleta: %s", out)
	}

	duration := parseDurationSec(parts[2])
	safeTitle := strings.ReplaceAll(parts[0], "/", "-")
	fileName := fmt.Sprintf("%s - %s.mp3", parts[1], safeTitle)
	filePath := filepath.Join("songs", fileName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		_, _, err := runYtDlp(
			"--cookies", "cookies.txt",
			"-f", "bestaudio",
			"-x", "--audio-format", "mp3",
			"-o", filePath,
			videoURL,
		)
		if err != nil {
			return nil, fmt.Errorf("yt-dlp download error: %w", err)
		}
	}

	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err == nil {
		defer tag.Close()
		tag.SetTitle(parts[0])
		tag.SetArtist(parts[1])
		tag.AddTextFrame(tag.CommonID("Length"), tag.DefaultEncoding(), fmt.Sprintf("%d", duration))
		if pic := GetAttPic(parts[3]); pic != nil {
			tag.AddAttachedPicture(*pic)
		}
		_ = tag.Save()
	}

	s := &Song{
		Title:     parts[0],
		Uploader:  parts[1],
		Duration:  duration,
		Thumbnail: parts[3],
		URL:       videoURL,
		Path:      filePath,
	}
	addSongToCache(*s)

	return s, nil
}

// GetAttPic descarga imagen y devuelve PictureFrame
func GetAttPic(url string) *id3v2.PictureFrame {
	resp, err := http.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	imgData := new(bytes.Buffer)
	if _, err := imgData.ReadFrom(resp.Body); err != nil {
		return nil
	}
	imgType := http.DetectContentType(imgData.Bytes())

	return &id3v2.PictureFrame{
		Encoding:    id3v2.EncodingUTF8,
		MimeType:    imgType,
		PictureType: id3v2.PTFrontCover,
		Description: "Cover",
		Picture:     imgData.Bytes(),
	}
}

func parseDurationSec(s string) time.Duration {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return time.Duration(n) * time.Second
}

// runYtDlp ejecuta yt-dlp
func runYtDlp(args ...string) (string, string, error) {
	var out, stderr bytes.Buffer
	cmd := exec.Command(YtDlpBin, args...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("[yt-dlp] Error: %v", err)
		log.Printf("[yt-dlp] stderr: %s", stderr.String())
	}
	return out.String(), stderr.String(), err
}
