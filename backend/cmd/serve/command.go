package serve

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Asheze1127/progress-checker/backend/util"
)

// Run starts the HTTP server with all dependencies wired.
func Run() error {
	cfg, err := util.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	router, err := wireRouter(cfg)
	if err != nil {
		return fmt.Errorf("failed to wire dependencies: %w", err)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Starting server on %s", addr)

	return http.ListenAndServe(addr, router)
}
