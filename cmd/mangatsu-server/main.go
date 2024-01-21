package main

import (
	"time"

	"github.com/Mangatsu/server/internal/config"
	"github.com/Mangatsu/server/pkg/api"
	"github.com/Mangatsu/server/pkg/cache"
	"github.com/Mangatsu/server/pkg/db"
	"github.com/Mangatsu/server/pkg/log"
	"github.com/Mangatsu/server/pkg/utils"
	"go.uber.org/zap"
)

func main() {
	config.LoadEnv()
	log.InitializeLogger(config.AppEnvironment, config.LogLevel)
	config.SetEnv()
	cache.InitPhysicalCache()
	db.InitDB()
	db.EnsureLatestVersion()

	username, password := config.GetInitialAdmin()
	users, err := db.GetUser(username)
	if err != nil {
		log.Z.Error("error fetching initial admin", zap.String("error", err.Error()))
	}

	if users == nil || len(users) == 0 {
		if err := db.Register(username, password, db.SuperAdmin); err != nil {
			log.Z.Fatal("error registering initial admin: ", zap.String("err", err.Error()))
		}
	}

	// Parse libraries from the environmental and insert/update to the db.
	libraries := config.ParseBasePaths()
	if err = db.StorePaths(libraries); err != nil {
		log.Z.Fatal("error saving library to db: ", zap.String("err", err.Error()))
	}

	cache.InitGalleryCache()
	cache.InitProcessingStatusCache()

	// Tasks
	utils.PeriodicTask(time.Minute, cache.PruneCache)
	utils.PeriodicTask(time.Minute, db.PruneExpiredSessions)

	api.LaunchAPI()
}
