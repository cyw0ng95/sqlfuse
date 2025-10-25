package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/generators/turso"
)

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
