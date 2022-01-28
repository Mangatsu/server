package main

import (
	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/api"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/library"
	log "github.com/sirupsen/logrus"
)

func main() {
	config.LoadEnv()
	library.InitCache()
	db.Initdb()
	db.EnsureLatestVersion()

	username, password := config.GetInitialAdmin()
	users, _ := db.GetUser(username)
	if users == nil || len(users) == 0 {
		err := db.Register(username, password, int64(db.Admin))
		if err != nil {
			log.Fatal("Error registering initial admin: ", err)
		}
	}

	// Parse libraries from the environmental and insert/update to the db.
	libraries := config.ParseBasePaths()
	db.StorePaths(libraries)

	// Scan the libraries for metadata and insert/update to the db.
	//startScan := time.Now()
	//library.ScanArchives()
	//metadata.ParseX()
	//metadata.ParseTitles(true, false)
	//elapsedScan := time.Since(startScan)
	//log.Info("Finding and tagging archives took: ", elapsedScan)

	// Generate thumbnails for all the archives, prioritizing covers.
	//startThumbnail := time.Now()
	//library.GenerateThumbnails(true)
	//elapsedThumbnail := time.Since(startThumbnail)
	//log.Info("Generating thumbnails took: ", elapsedThumbnail)

	api.LaunchAPI()
}
