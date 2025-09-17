package player

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	id3v2 "github.com/bogem/id3v2"
)

// --- Constantes ---
const (
	YtDlpBin = "yt-dlp"
)

// runYtDlp es un wrapper para ejecutar yt-dlp con los argumentos dados.
// Devuelve stdout, stderr y error. Además loguea para debug.
func runYtDlp(args ...string) (string, string, error) {
	var out bytes.Buffer
	var stderr bytes.Buffer

	// Prepend -vU a los args del usuario

	cmd := exec.Command(YtDlpBin, args...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

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


// YtdlpSearch busca en YouTube y devuelve resultados JSON.
func YtdlpSearch(query string, limit int) ([]string, error) {
	if limit <= 0 || limit >=5 {
		limit = 5
	}
	out, stderr, err := runYtDlp("--dump-json", "--flat-playlist", fmt.Sprintf("ytsearch%d:%s", limit, query))
	if err != nil {
		return nil, fmt.Errorf("yt-dlp search error: %w - %s", err, stderr)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	return lines, nil
}


// Metadata representa lo que nos interesa de yt-dlp
type Metadata struct {
	Title     string
	Uploader  string
	Album     string
	Duration  time.Duration
	Thumbnail string
}

func parseDurationSec(s string) time.Duration {
    n, err := strconv.Atoi(strings.TrimSpace(s))
    if err != nil || n <= 0 {
        return 0 // desconocido
    }
    return time.Duration(n) * time.Second
}

// YtdlpBestAudioURL descarga y embebe metadatos en el MP3
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

	meta := Metadata{
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

	if err = tag.Save(); err != nil {
		return "", fmt.Errorf("error guardando metadatos: %w", err)
	}

	log.Printf("[yt-dlp] Descarga + metadata completada: %s", filePath)
	return "/songs/" + fileName, nil
}

