package photoprism

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

var backupAlbumsLatest = time.Time{}
var backupAlbumsMutex = sync.Mutex{}

// BackupAlbums creates a YAML file backup of all albums.
func BackupAlbums(backupPath string, force bool) (count int, err error) {
	// Make sure only one backup/restore operation is running at a time.
	backupAlbumsMutex.Lock()
	defer backupAlbumsMutex.Unlock()

	albums, queryErr := query.Albums(0, 1000000)

	if queryErr != nil {
		return count, queryErr
	}

	if !fs.PathExists(backupPath) {
		backupPath = Config().BackupAlbumsPath()
	}

	log.Infof("exporting album metadata to YAML backup files")
	log.Debugf("album backups will be created in %s", clean.Log(backupPath))

	var latest time.Time

	if !force {
		latest = backupAlbumsLatest
	}

	for _, a := range albums {
		if !force && a.UpdatedAt.Before(backupAlbumsLatest) {
			continue
		}

		if a.UpdatedAt.After(latest) {
			latest = a.UpdatedAt
		}

		fileName := a.YamlFileName(backupPath)

		if saveErr := a.SaveAsYaml(fileName); saveErr != nil {
			log.Errorf("album: %s (save as yaml)", saveErr)
			err = saveErr
		} else {
			log.Tracef("album: updated backup file %s", clean.Log(filepath.Base(fileName)))
			count++
		}
	}

	backupAlbumsLatest = latest

	return count, err
}
