package scanner

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tnm-yes-anyidea/auD-limit/internal/db"
)

var AudioExts = map[string]bool{
	".mp3": true, 
	".flac": true,
	".ogg": true,
	".m4a": true,
	".aac": true,
	".wma": true,
	".alac": true,
	".aiff": true,
	".m4b": true,
	".opus": true,
	".wav": true,
	".ape": true,
	".wv": true,
}

// ffprobe format JSON structure (partial)
type ffprobeFormat struct {
	Tags map[string]string `json:"tags"`
	Duration string `json:"duration"`
}

type ffprobeOut struct{
	Format ffprobeFormat `json:"format"`
}

func ffprobeTags(path string) (map[string]string, error) {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", path)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var f ffprobeOut
	if err := json.Unmarshal(out, &f); err != nil { return nil, err }
	m := make(map[string]string)
	for k,v := range f.Format.Tags { m[strings.ToLower(k)] = v }
	return m, nil
}

func ffprobeDuration(path string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", path)
	out, err := cmd.Output()
	if err != nil { return 0, err }
	s := strings.TrimSpace(string(out))
	if s == "" { return 0, nil }
	var v float64
	fmt.Sscanf(s, "%f", &v)
	return v, nil
}

func ScanAndStore(d *db.DB, root string) error {
	return filepath.Walk(root, func(p string, info filepath.FileInfo, err error) error {
		if err != nil { return nil }
		if info.IsDir() { return nil }
		ext := strings.ToLower(filepath.Ext(p))
		if !AudioExts[ext] { return nil }
		// try get tags/duration
		tags, _ := ffprobeTags(p)
		dur, _ := ffprobeDuration(p)
		t := db.Track{
			Path: p,
			Title: firstNonEmpty(tags["title"], filepath.Base(strings.TrimSuffix(p, ext))),
			Artist: tags["artist"],
			Album: tags["album"],
			Duration: dur,
		}
		if err := d.UpsertTrack(t); err != nil {
			return err
		}
		return nil
	})
}

func firstNonEmpty(a, b string) string { if strings.TrimSpace(a) != "" { return a }; return b }