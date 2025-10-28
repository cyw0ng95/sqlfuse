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

// ServerConfig holds server-wide configuration
type ServerConfig struct {
	Port                string    `json:"port"`
	ExecutorsConfigPath string    `json:"executors_config_path"`
	ServerName          string    `json:"server_name"`
	ServerVersion       string    `json:"server_version"`
	Job                 JobConfig `json:"job"`
}

// JobConfig holds job-related configuration
type JobConfig struct {
	MaxOutputBytes int    `json:"max_output_bytes"`
	PersistPath    string `json:"persist_path"`
}

var serverConfig ServerConfig
var executorsConfig []ExecutorConfig
var executorsByName map[string]executors.Executor

func loadServerConfig(path string) (ServerConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return ServerConfig{}, err
	}
	var cfg ServerConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return ServerConfig{}, err
	}
	return cfg, nil
}

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

	// Load server configuration
	serverCfgPath := os.Getenv("SERVER_CONFIG_PATH")
	if serverCfgPath == "" {
		serverCfgPath = "./config/server.json"
	}
	cfg, err := loadServerConfig(serverCfgPath)
	if err != nil {
		common.Logger.Warn().Err(err).Msgf("failed to load server config from %s, using defaults", serverCfgPath)
		// Set default values
		cfg = ServerConfig{
			Port:                "8080",
			ExecutorsConfigPath: "./config/executors.json",
			ServerName:          "sqlsmith-go minimal server",
			ServerVersion:       "0.1",
			Job: JobConfig{
				MaxOutputBytes: 65536, // 64KiB
				PersistPath:    "./output/jobs.json",
			},
		}
	}
	serverConfig = cfg

	// Allow PORT environment variable to override config
	if envPort := os.Getenv("PORT"); envPort != "" {
		serverConfig.Port = envPort
	}

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
			"name":    serverConfig.ServerName,
			"version": serverConfig.ServerVersion,
		})
	})

	// Generator metadata endpoint
	e.GET("/generators/get", func(c echo.Context) error {
		// Create a temporary generator instance to get metadata
		gen := turso.NewGenerator(0)
		resp := map[string]interface{}{
			"generator": gen.Name(),
			"stmts":     gen.SupportedStmts(),
			"note":      "These are the high-level statement types the generator can produce with their default weights.",
		}
		return c.JSON(http.StatusOK, resp)
	})

	// Load executors configuration
	if execCfg, err := loadExecutorsConfig(serverConfig.ExecutorsConfigPath); err != nil {
		common.Logger.Warn().Err(err).Msgf("failed to load executors config from %s", serverConfig.ExecutorsConfigPath)
	} else {
		executorsConfig = execCfg
		executorsByName = buildExecutorsMap(execCfg)
		common.Logger.Info().Msgf("Loaded %d executors from config", len(execCfg))
	}

	// Expose executors list via HTTP
	e.GET("/executors", func(c echo.Context) error {
		return c.JSON(http.StatusOK, executorsConfig)
	})

	// Register job routes (moved to job_ctrl.go)
	RegisterJobRoutes(e)

	common.Logger.Info().Msgf("Starting minimal server on port %s", serverConfig.Port)
	if err := e.Start(":" + serverConfig.Port); err != nil {
		fmt.Fprintf(os.Stderr, "server failed to start: %v\n", err)
		os.Exit(1)
	}
}
