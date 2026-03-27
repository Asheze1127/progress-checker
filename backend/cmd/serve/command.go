package serve

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/Asheze1127/progress-checker/backend/util"
)

// Run starts the HTTP server with all dependencies wired.
func Run() error {
	gin.SetMode(gin.ReleaseMode)

	cfg, err := util.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logLevel := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))

	router, err := wireRouter(cfg)
	if err != nil {
		return fmt.Errorf("failed to wire dependencies: %w", err)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	slog.Info("starting server", slog.String("addr", addr))

	return http.ListenAndServe(addr, router)
}
