package infra

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"feints/internal/infra/core"
)

const (
	YtDlpBin = "yt-dlp"
)

func run(args ...string) (string, string, error) {
	cmd := exec.Command(YtDlpBin, args...)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	return out.String(), stderr.String(), err
}

func Search(query string, limit int) ([]core.Song, error) {
	if limit <= 0 || limit > 5 {
		limit = 5
	}

	out, stderr, err := run(
		"--cookies", "cookies.txt",
		"--dump-json",
		fmt.Sprintf("ytsearch%d:music%s", limit, query),
	)
	if err != nil {
		return nil, fmt.Errorf("yt-dlp search error: %w - %s", err, stderr)
	}

	lines := bytes.Split([]byte(out), []byte("\n"))
	var results []core.Song
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal(line, &raw); err != nil {
			continue
		}

		s := core.Song{
			Title:     fmt.Sprint(raw["title"]),
			Uploader:  fmt.Sprint(raw["uploader"]),
			Thumbnail: fmt.Sprint(raw["thumbnail"]),
			URL:       fmt.Sprint(raw["webpage_url"]),
		}
		if dur, ok := raw["duration"].(float64); ok {
			s.Duration = time.Duration(int(dur)) * time.Second
		}
		results = append(results, s)
	}
	return results, nil
}

func Metadata(url string) (*core.Song, error) {
	out, stderr, err := run("--cookies", "cookies.txt", "--dump-single-json", url)
	if err != nil {
		return nil, fmt.Errorf("yt-dlp metadata error: %w - %s", err, stderr)
	}

	var raw map[string]any
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		return nil, fmt.Errorf("json parse error: %w", err)
	}

	s := &core.Song{
		Title:     fmt.Sprint(raw["title"]),
		Uploader:  fmt.Sprint(raw["uploader"]),
		Thumbnail: fmt.Sprint(raw["thumbnail"]),
		URL:       url,
	}
	if dur, ok := raw["duration"].(float64); ok {
		s.Duration = time.Duration(int(dur)) * time.Second
	}
	return s, nil
}

func DownloadAudio(url, path string) error {
	_, stderr, err := run(
		"--cookies", "cookies.txt",
		"-x", "--audio-format", "mp3",
		"--output", path,
		url,
	)
	if err != nil {
		return fmt.Errorf("yt-dlp download error: %w - %s", err, stderr)
	}
	return nil
}
