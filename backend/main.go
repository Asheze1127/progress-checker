package main

import (
	"log"

	"github.com/Asheze1127/progress-checker/backend/cmd/serve"
)

func main() {
	if err := serve.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
