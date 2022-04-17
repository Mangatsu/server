package cache

import (
	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/library"
	"github.com/djherbis/atime"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"sync"
	"time"
)

type cacheValue struct {
	Accessed time.Time
	Mu       *sync.Mutex
}

type GalleryCache struct {
	Path  string
	Store map[string]cacheValue
}

var galleryCache *GalleryCache

// Init initializes the abstraction layer for the gallery cache.
func Init() {
	galleryCache = &GalleryCache{
		Path:  config.BuildCachePath(),
		Store: make(map[string]cacheValue),
	}

	iterateCacheEntries(func(pathToEntry string, accessTime time.Time) {
		maybeUUID := path.Base(pathToEntry)
		if _, err := uuid.Parse(maybeUUID); err != nil {
			return
		}

		galleryCache.Store[path.Base(pathToEntry)] = cacheValue{
			Accessed: accessTime,
			Mu:       &sync.Mutex{},
		}
	})
}

// PruneCache removes entries not accessed (internal timestamp in mem) in the last x time in a thread-safe manner.
func PruneCache() {
	now := time.Now()
	for galleryUUID, value := range galleryCache.Store {
		value.Mu.Lock()
		if value.Accessed.Add(config.Options.Cache.TTL).Before(now) {
			if err := remove(galleryUUID); err != nil {
				log.Errorf("Error occured while deleting cache entry: %s", err)
			}
		}
		value.Mu.Unlock()
	}
}

// PruneCacheFS removes entries not accessed (filesystem timestamp) in the last x time. Not thread-safe.
func PruneCacheFS() {
	now := time.Now()
	iterateCacheEntries(func(pathToEntry string, accessTime time.Time) {
		if accessTime.Add(config.Options.Cache.TTL).Before(now) {
			if err := os.Remove(pathToEntry); err != nil {
				log.Errorf("Error occured while removing cache entry: %s", err)
			}
		}
	})
}

// Read reads the cached gallery from disk. If it doesn't exist, it will be created and then read.
func Read(archivePath string, galleryUUID string) ([]string, int) {
	galleryCache.Store[galleryUUID] = cacheValue{
		Accessed: time.Now(),
		Mu:       &sync.Mutex{},
	}

	galleryCache.Store[galleryUUID].Mu.Lock()
	defer galleryCache.Store[galleryUUID].Mu.Unlock()

	files, count := library.ReadGallery(archivePath, galleryUUID)
	if count == 0 {
		return files, count
	}

	return files, count
}

// remove wipes the cached gallery from disk.
func remove(galleryUUID string) error {
	// Paranoid check to make sure that the base is a real UUID, since we don't want to delete anything else.
	maybeUUID := path.Base(galleryUUID)
	if _, err := uuid.Parse(maybeUUID); err != nil {
		delete(galleryCache.Store, galleryUUID)
		return err
	}

	galleryPath := config.BuildCachePath(galleryUUID)
	if err := os.RemoveAll(galleryPath); err != nil {
		if os.IsNotExist(err) {
			delete(galleryCache.Store, galleryUUID)
		}
		return err
	}

	delete(galleryCache.Store, galleryUUID)

	return nil
}

// iterateCacheEntries iterates over all cache entries and calls the callback function for each entry.
func iterateCacheEntries(callback func(pathToEntry string, accessTime time.Time)) {
	cachePath := config.BuildCachePath()
	cacheEntries, err := os.ReadDir(cachePath)
	if err != nil {
		log.Errorf("Could not read cache dir: %s", err)
		return
	}

	for _, entry := range cacheEntries {
		info, err := entry.Info()
		if err != nil {
			log.Errorf("Error occured while reading cache entry info: %s", err)
			return
		}

		pathToEntry := path.Join(cachePath, entry.Name())
		accessTime, err := atime.Stat(pathToEntry)
		if err != nil {
			log.Debugf("Could not read the access time of '%s'.  %s", entry.Name(), err)
			accessTime = info.ModTime()
		}

		callback(pathToEntry, accessTime)
	}
}
