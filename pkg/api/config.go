package api

type Config struct {
	MongodbURI string
	Port       int
	ProjectID  string
}

func NewConfig(mongodbURI string, port int, projectID string) Config {
	return Config{
		MongodbURI: mongodbURI,
		Port:       port,
		ProjectID:  projectID,
	}
}
