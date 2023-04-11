package engine

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type Engine struct {
	Config
	Users      *mongo.Collection
	Challenges *mongo.Collection
	Activities *mongo.Collection
}

type Config struct {
	mongodb_uri string
	port        int
}

func NewConfig(mongodb_uri string, port int) Config {
	return Config{
		mongodb_uri: mongodb_uri,
		port:        port,
	}
}

func NewEngine(c Config, users *mongo.Collection, challenges *mongo.Collection, activities *mongo.Collection) *Engine {
	return &Engine{
		c,
		users,
		challenges,
		activities,
	}
}
