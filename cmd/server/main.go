package main

import (
	"fmt"
	"os"
)

func main() {
	// ...
	fmt.Println("Creds:", os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))

	// If DEV, load config from .env
	// If STG / PROD, load config from google cloud secret manager
}
