package main

import (
	"log"
	"nfs002/template/v1/api"
	"nfs002/template/v1/internal/db"
	"nfs002/template/v1/internal/utils"
	"os"
)

func main() {

	// Load migrations
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime|log.Lshortfile)
	if utils.GetBoolEnvOrDefault("RUN_MIGRAGTIONS", false) {
		if err := db.MigrateUp(); err != nil {
			errorLog.Fatal(err)
		} else {
			infoLog.Printf("succesfully ran migrations")
		}
	}
	// Run API
	api.Run()
}
