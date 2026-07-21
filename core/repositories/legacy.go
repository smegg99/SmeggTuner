package repositories

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"smegg.me/smeggtuner/core/session"
)

// LegacyStats records what the one-time import brought forward and skipped.
type LegacyStats struct {
	Sessions    int
	Instruments int
	Skipped     int
}

// Empty reports whether the import found nothing to do.
func (s LegacyStats) Empty() bool {
	return s.Sessions == 0 && s.Instruments == 0 && s.Skipped == 0
}

// ImportLegacy brings a pre-database data directory (sessions/*.session.json, instruments/*.json+.jpg) into the datastore, but only when it is empty, so re-import can't resurrect deleted rows; old files are left in place, unparseable ones skipped.
func ImportLegacy(dir string) (LegacyStats, error) {
	var stats LegacyStats

	var sessions, templates int64
	if err := db().Model(&session.Session{}).Count(&sessions).Error; err != nil {
		return stats, err
	}
	if err := db().Model(&session.Template{}).Count(&templates).Error; err != nil {
		return stats, err
	}
	if sessions > 0 || templates > 0 {
		return stats, nil
	}

	repo := GetSessionRepository()
	for _, path := range legacyFiles(filepath.Join(dir, "sessions"), session.LegacyFileExt) {
		s, err := session.Read(path)
		if err != nil {
			stats.Skipped++
			continue
		}
		if err := repo.Insert(s); err != nil {
			stats.Skipped++
			continue
		}
		stats.Sessions++
	}

	instruments := GetInstrumentRepository()
	for _, path := range legacyFiles(filepath.Join(dir, "instruments"), ".json") {
		t, err := session.ReadTemplate(path)
		if err != nil || !session.ValidID(t.ID) {
			stats.Skipped++
			continue
		}
		row := session.Template{ID: t.ID, Name: t.Name, Instrument: t.Instrument}
		if jpg, err := os.ReadFile(strings.TrimSuffix(path, ".json") + ".jpg"); err == nil {
			row.Image = jpg
			row.HasImage = true
			row.ImageRev = time.Now().UnixNano()
		}
		if err := instruments.Create(&row); err != nil {
			stats.Skipped++
			continue
		}
		stats.Instruments++
	}

	return stats, nil
}

// legacyFiles lists files with ext in dir; a missing dir yields nothing.
func legacyFiles(dir, ext string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ext) {
			continue
		}
		out = append(out, filepath.Join(dir, e.Name()))
	}
	return out
}
