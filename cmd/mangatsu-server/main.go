package main

import (
	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/api"
	"github.com/Mangatsu/server/pkg/cache"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/library"
	"github.com/Mangatsu/server/pkg/utility"
	log "github.com/sirupsen/logrus"
	"time"
)

func main() {
	config.LoadEnv()
	library.InitCache()
	db.Initdb()
	db.EnsureLatestVersion()

	username, password := config.GetInitialAdmin()
	users, err := db.GetUser(username)
	if err != nil {
		log.Error(err)
	}

	if users == nil || len(users) == 0 {
		if err := db.Register(username, password, int64(db.Admin)); err != nil {
			log.Fatal("Error registering initial admin: ", err)
		}
	}

	// Parse libraries from the environmental and insert/update to the db.
	libraries := config.ParseBasePaths()
	if err = db.StorePaths(libraries); err != nil {
		log.Fatal("Error saving library to db: ", err)
	}

	cache.Init()

	// Tasks
	utility.PeriodicTask(time.Minute, cache.PruneCache)
	utility.PeriodicTask(time.Minute, db.PruneExpiredSessions)

	api.LaunchAPI()
}
