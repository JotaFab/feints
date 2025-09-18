package player

import (
	"bytes"
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

var (
	searchCache   = make(map[string][]string)
	searchCacheMu sync.RWMutex
	searchLock    sync.Mutex
	searchTTL     = 10 * time.Minute
	cacheTime     = make(map[string]time.Time)
	cnt int
)


// --- Constantes ---
const (
	YtDlpBin = "yt-dlp"
)


func YtdlpSearch(query string, limit int) ([]string, error) {
	if limit <= 0 || limit >= 5 {
		limit = 5
	}

	// Revisar cache primero
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

	// Re-chequear cache después de obtener lock
	searchCacheMu.RLock()
	cached, ok = searchCache[query]
	timestamp, tsOk = cacheTime[query]
	searchCacheMu.RUnlock()
	if ok && tsOk && time.Since(timestamp) < searchTTL {
		return cached, nil
	}


	out, stderr, err := runYtDlp(
		"--dump-json",
		"--flat-playlist",
		fmt.Sprintf("ytsearch%d:%s", limit, query),
	)
	if err != nil {
		return nil, fmt.Errorf("yt-dlp search error: %w - %s", err, stderr)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")

	// Guardar en cache
	searchCacheMu.Lock()
	searchCache[query] = lines
	cacheTime[query] = time.Now()
	log.Println("[yt-dlp] Cache actualizado:")
	for k, v := range searchCache {
		log.Printf("  Query: %s -> %d resultados", k, len(v))
	}
	searchCacheMu.Unlock()

	return lines, nil
}


// runYtDlp es un wrapper para ejecutar yt-dlp con los argumentos dados.
// Devuelve stdout, stderr y error. Además loguea para debug.
func runYtDlp(args ...string) (string, string, error) {
	var out bytes.Buffer
	var stderr bytes.Buffer

	// Prepend -vU a los args del usuario
	

	cmd := exec.Command(YtDlpBin, args...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	// Ejecutar yt-dlp search
	cnt++
	log.Printf("counter of searchs: %d\n", cnt)
	log.Printf("[yt-dlp] Ejecutando: %s %s", YtDlpBin, strings.Join(args, " "))

	err := cmd.Run()
	if err != nil {
		log.Printf("[yt-dlp] Error: %v", err)
		log.Printf("[yt-dlp] stderr: %s", stderr.String())
		log.Printf("[yt-dlp] stdout parcial: %s", out.String())
	}

	return out.String(), stderr.String(), err
}

// YtdlpVersion devuelve la versión instalada de yt-dlp.
func YtdlpVersion() (string, error) {
	out, stderr, err := runYtDlp("--version")
	if err != nil {
		return "", fmt.Errorf("yt-dlp version error: %v - %s", err, stderr)
	}
	return strings.TrimSpace(out), nil
}

func parseDurationSec(s string) time.Duration {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n <= 0 {
		return 0 // desconocido
	}
	return time.Duration(n) * time.Second
}

// YtdlpBestAudioURL descarga y embebe metadatos en el MP3 y devuelve la ruta de la cancion descargada
func YtdlpBestAudioURL(videoURL string) (string, error) {
	// 1. Obtener metadata mínima con yt-dlp
	out, _, err := runYtDlp(
		"--cookies", "cookies.txt",
		"--flat-playlist",
		"--print", "%(title)s|%(uploader)s|%(duration)s|%(thumbnail)s",
		videoURL,
	)
	if err != nil {
		return "", fmt.Errorf("error obteniendo metadata: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(out), "|")
	if len(parts) < 4 {
		return "", fmt.Errorf("metadata incompleta: %s", out)
	}
	// Metadata representa lo que nos interesa de yt-dlp
	type metadata struct {
		Title     string
		Uploader  string
		Album     string
		Duration  time.Duration
		Thumbnail string
	}
	meta := metadata{
		Title:     parts[0],
		Uploader:  parts[1],
		Duration:  parseDurationSec(parts[2]),
		Thumbnail: parts[3],
	}

	// Nombre de archivo seguro
	safeTitle := strings.ReplaceAll(meta.Title, "/", "-")
	fileName := fmt.Sprintf("%s - %s.mp3", meta.Uploader, safeTitle)
	filePath := filepath.Join("songs", fileName)

	// 2. Verificar si ya existe
	if _, err := os.Stat(filePath); err == nil {
		log.Printf("[yt-dlp] Archivo ya existe, no se descarga: %s", filePath)
		return "songs/" + fileName, nil
	}

	// 3. Descargar audio como MP3
	args := []string{
		"--cookies", "cookies.txt",
		"-f", "bestaudio",
		"-x", "--audio-format", "mp3",
		"-o", filePath,
		videoURL,
	}
	if _, _, err := runYtDlp(args...); err != nil {
		return "", fmt.Errorf("yt-dlp download error: %w", err)
	}

	// 4. Escribir metadatos ID3 en el MP3
	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return "", fmt.Errorf("error abriendo MP3 para metadata: %w", err)
	}
	defer tag.Close()

	tag.SetTitle(meta.Title)
	tag.SetArtist(meta.Uploader)
	if meta.Album != "" {
		tag.SetAlbum(meta.Album)
	}
	if meta.Duration > 0 {
		tag.AddTextFrame(tag.CommonID("Length"), tag.DefaultEncoding(), fmt.Sprintf("%d", meta.Duration))
	}

	// TODO: Descargar e incrustar la carátula desde meta.Thumbnail
	// (se requiere hacer un GET y usar tag.AddAttachedPicture)
	attPic := GetAttPic(meta.Thumbnail)
	tag.AddAttachedPicture(*attPic)
	if err = tag.Save(); err != nil {
		return "", fmt.Errorf("error guardando metadatos: %w", err)
	}

	log.Printf("[yt-dlp] Descarga + metadata completada: %s", filePath)
	return "songs/" + fileName, nil
}

// GetAttPic descarga la imagen desde la URL y devuelve un AttachedPictureFrame
func GetAttPic(url string) *id3v2.PictureFrame {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error descargando imagen:", err)
		return nil
	}
	defer resp.Body.Close()

	imgData := new(bytes.Buffer)
	if _, err := imgData.ReadFrom(resp.Body); err != nil {
		fmt.Println("Error leyendo imagen:", err)
		return nil
	}

	// Detectar tipo MIME automáticamente
	imgType := http.DetectContentType(imgData.Bytes())

	return &id3v2.PictureFrame{
		Encoding:    id3v2.EncodingUTF8,
		MimeType:    imgType,
		PictureType: id3v2.PTFrontCover,
		Description: "Cover",
		Picture:     imgData.Bytes(),
	}
}

