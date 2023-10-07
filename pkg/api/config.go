package api

import "github.com/AustinBayley/activity_tracker_api/pkg/service"

type Config struct {
	Environment Environment
	MongodbURI  string
	DBName      string
	Port        int
	ProjectID   string
	MapsAPIKey  string
	UserContext service.RequestContext
}

func NewConfig(environment Environment, mongodbURI string, dbName string, port int, projectID string, mapsAPIKey string, userContext service.RequestContext) Config {
	return Config{
		Environment: environment,
		MongodbURI:  mongodbURI,
		DBName:      dbName,
		Port:        port,
		ProjectID:   projectID,
		MapsAPIKey:  mapsAPIKey,
		UserContext: userContext,
	}
}
