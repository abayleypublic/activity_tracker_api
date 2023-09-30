package api

type Config struct {
	Environment Environment
	MongodbURI  string
	DBName      string
	Port        int
	ProjectID   string
	MapsAPIKey  string
}

func NewConfig(environment Environment, mongodbURI string, dbName string, port int, projectID string, mapsAPIKey string) Config {
	return Config{
		Environment: environment,
		MongodbURI:  mongodbURI,
		DBName:      dbName,
		Port:        port,
		ProjectID:   projectID,
		MapsAPIKey:  mapsAPIKey,
	}
}
