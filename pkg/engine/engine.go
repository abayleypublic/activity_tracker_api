package engine

type Engine struct {
	Config
}

type Config struct {
	mongodb_uri string
}

func NewConfig(mongodb_uri string) Config {
	return Config{
		mongodb_uri: mongodb_uri,
	}
}

func NewEngine(c Config) *Engine {
	return &Engine{
		c,
	}
}
