package main

import (
	"log"
	"os"
	"strconv"

	"github.com/AustinBayley/activity_tracker_api/pkg/api"
	"github.com/AustinBayley/activity_tracker_api/pkg/service"
)

const (
	projectID string = "portfolio-459420"
	dbName    string = "roam"
)

func main() {
	var env api.Environment
	if value, ok := os.LookupEnv("ENVIRONMENT"); ok {
		env = api.Environment(value)
	} else {
		env = api.DEV
	}

	mapsKey := os.Getenv("MAPS_KEY")
	mongoURI := os.Getenv("MONGODB_URI")

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatalln(err)
	}

	user := os.Getenv("USER")
	admin, err := strconv.ParseBool(os.Getenv("ADMIN"))
	if err != nil {
		admin = false
	}

	cfg := api.NewConfig(env, mongoURI, dbName, port, projectID, mapsKey, service.RequestContext{
		UserID: service.ID(user),
		Admin:  admin,
	})

	a, err := api.NewAPI(cfg)

	if err != nil {
		log.Fatalln(err)
	}

	a.Start()
}
