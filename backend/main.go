package main

import (
	"log/slog"
	"os"

	"github.com/Asheze1127/progress-checker/backend/cmd/serve"
)

func main() {
	if err := serve.Run(); err != nil {
		slog.Error("server error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
