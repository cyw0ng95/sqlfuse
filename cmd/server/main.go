package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/executors"
	"sqlsmith-go/internal/generators/turso"
)

// ExecutorConfig describes an executable exposed to the server via config.
type ExecutorConfig struct {
	Executor string `json:"executor"`
	Path     string `json:"path"`
}

var executorsConfig []ExecutorConfig
var executorsByName map[string]executors.Executor

func loadExecutorsConfig(path string) ([]ExecutorConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg []ExecutorConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func buildExecutorsMap(cfg []ExecutorConfig) map[string]executors.Executor {
	m := make(map[string]executors.Executor)
	for _, e := range cfg {
		// resolve path and verify
		abs, _ := filepath.Abs(e.Path)
		if p, err := executors.ResolveExecPath(filepath.Dir(abs), filepath.Base(abs)); err == nil {
			m[e.Executor] = executors.NewCmdExecutor(e.Executor, p)
		} else {
			common.Logger.Warn().Err(err).Msgf("skipping executor %s (path=%s)", e.Executor, e.Path)
		}
	}
	return m
}

func main() {
	common.InitLogger()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	e.GET("/info", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"name":    "sqlsmith-go minimal server",
			"version": "0.1",
		})
	})

	// Generator metadata endpoint
	e.GET("/generators/get", func(c echo.Context) error {
		gen := turso.Info
		resp := map[string]interface{}{
			"generator": gen.Name(),
			"stmts":     gen.SupportedStmts(),
			"note":      "These are the high-level statement types the generator can produce with their default weights.",
		}
		return c.JSON(http.StatusOK, resp)
	})

	// Load executors configuration
	cfgPath := "/opt/config/executors.json"
	if cfg, err := loadExecutorsConfig(cfgPath); err != nil {
		common.Logger.Warn().Err(err).Msgf("failed to load executors config from %s", cfgPath)
	} else {
		executorsConfig = cfg
		executorsByName = buildExecutorsMap(cfg)
		common.Logger.Info().Msgf("Loaded %d executors from config", len(cfg))
	}

	// Expose executors list via HTTP
	e.GET("/executors", func(c echo.Context) error {
		return c.JSON(http.StatusOK, executorsConfig)
	})

	// Register job routes (moved to job_ctrl.go)
	RegisterJobRoutes(e)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	common.Logger.Info().Msgf("Starting minimal server on port %s", port)
	if err := e.Start(":" + port); err != nil {
		fmt.Fprintf(os.Stderr, "server failed to start: %v\n", err)
		os.Exit(1)
	}
}
