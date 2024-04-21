package main

import (
	"nfs002/template/v1/api"
	"nfs002/template/v1/internal/db"
	"nfs002/template/v1/internal/utils"
	u "nfs002/template/v1/internal/utils"
)

func main() {
	// Load environment
	u.LoadEnv()

	// Load migrations
	if utils.GetBoolEnvOrDefault("RUN_MIGRAGTIONS", false) {
		if err := db.MigrateUp(); err != nil {
			u.PanicLog("Failed to run migrations", err)
		} else {
			u.InfoLog("Succesfully ran migrations")
		}
	}

	api.Run()
}
