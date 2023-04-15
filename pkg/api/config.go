package api

type Config struct {
	MongodbUri string
	Port       int
}

func NewConfig(mongodbUri string, port int) Config {
	return Config{
		MongodbUri: mongodbUri,
		Port:       port,
	}
}
