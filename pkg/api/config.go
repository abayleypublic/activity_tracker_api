package api

type Config struct {
	MongodbURI string
	DBName     string
	Port       int
	ProjectID  string
	MapsAPIKey string
}

func NewConfig(mongodbURI string, dbName string, port int, projectID string, mapsAPIKey string) Config {
	return Config{
		MongodbURI: mongodbURI,
		DBName:     dbName,
		Port:       port,
		ProjectID:  projectID,
		MapsAPIKey: mapsAPIKey,
	}
}
