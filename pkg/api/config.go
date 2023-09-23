package api

type Config struct {
	MongodbURI string
	DBName     string
	Port       int
	ProjectID  string
}

func NewConfig(mongodbURI string, dbName string, port int, projectID string) Config {
	return Config{
		MongodbURI: mongodbURI,
		DBName:     dbName,
		Port:       port,
		ProjectID:  projectID,
	}
}
