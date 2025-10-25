package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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

	// --- Job manager implementation ---
	// Simple in-memory job store. Jobs invoke configured executables under ./output and capture stdout/stderr.
	type JobStatus string
	const (
		JobPending JobStatus = "pending"
		JobRunning JobStatus = "running"
		JobDone    JobStatus = "done"
		JobFailed  JobStatus = "failed"
		JobStopped JobStatus = "stopped"
	)

	type Job struct {
		ID        string     `json:"id"`
		Cmd       string     `json:"cmd"`
		Status    JobStatus  `json:"status"`
		PID       int        `json:"pid,omitempty"`
		ExitCode  *int       `json:"exit_code,omitempty"`
		StartedAt *time.Time `json:"started_at,omitempty"`
		EndedAt   *time.Time `json:"ended_at,omitempty"`

		Stdout bytes.Buffer `json:"-"`
		Stderr bytes.Buffer `json:"-"`

		cmd *exec.Cmd `json:"-"`
	}

	var (
		jobStore     = make(map[string]*Job)
		jobMu        sync.Mutex
		jobIDCounter uint64
		// max bytes to return for stdout/stderr in /job/info
		maxOutputReturn = 64 * 1024 // 64KB
	)

	// helper to truncate output for responses
	truncate := func(b []byte, n int) string {
		if len(b) <= n {
			return string(b)
		}
		return string(b[:n]) + "\n...(truncated)"
	}

	// POST /job/new { "executor": "name", "args": ["--flag","val"] }
	e.POST("/job/new", func(c echo.Context) error {
		var req struct {
			Executor string   `json:"executor"`
			Args     []string `json:"args"`
		}
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		if req.Executor == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "executor is required"})
		}

		exe, ok := executorsByName[req.Executor]
		if !ok {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "unknown executor"})
		}

		id := strconv.FormatUint(atomic.AddUint64(&jobIDCounter, 1), 10)
		fullCmd := exe.Path()
		if len(req.Args) > 0 {
			fullCmd = fullCmd + " " + strings.Join(req.Args, " ")
		}
		job := &Job{
			ID:     id,
			Cmd:    fullCmd,
			Status: JobPending,
		}

		jobMu.Lock()
		jobStore[id] = job
		jobMu.Unlock()

		// start the job asynchronously
		go func(j *Job, exe executors.Executor, args []string) {
			j.Status = JobRunning
			now := time.Now()
			j.StartedAt = &now

			cmd := exe.BuildCmd(args)
			j.cmd = cmd
			cmd.Stdout = &j.Stdout
			cmd.Stderr = &j.Stderr

			if err := cmd.Start(); err != nil {
				j.Status = JobFailed
				end := time.Now()
				j.EndedAt = &end
				return
			}

			if cmd.Process != nil {
				j.PID = cmd.Process.Pid
			}

			// wait for completion
			err := cmd.Wait()
			end := time.Now()
			j.EndedAt = &end
			if err != nil {
				// try to extract exit code
				if cmd.ProcessState != nil {
					exit := cmd.ProcessState.ExitCode()
					j.ExitCode = &exit
				}
				if j.Status != JobStopped {
					j.Status = JobFailed
				}
				return
			}

			if cmd.ProcessState != nil {
				exit := cmd.ProcessState.ExitCode()
				j.ExitCode = &exit
			}
			j.Status = JobDone
		}(job, exe, req.Args)

		return c.JSON(http.StatusAccepted, map[string]string{"id": id})
	})

	// GET /job/status?id=123
	e.GET("/job/status", func(c echo.Context) error {
		id := c.QueryParam("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
		}
		jobMu.Lock()
		job, ok := jobStore[id]
		jobMu.Unlock()
		if !ok {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "job not found"})
		}

		resp := map[string]interface{}{
			"id":     job.ID,
			"cmd":    job.Cmd,
			"status": job.Status,
			"pid":    job.PID,
		}
		if job.StartedAt != nil {
			resp["started_at"] = job.StartedAt.Format(time.RFC3339)
		}
		if job.EndedAt != nil {
			resp["ended_at"] = job.EndedAt.Format(time.RFC3339)
		}
		if job.ExitCode != nil {
			resp["exit_code"] = *job.ExitCode
		}
		return c.JSON(http.StatusOK, resp)
	})

	// POST /job/stop?id=123
	e.POST("/job/stop", func(c echo.Context) error {
		id := c.QueryParam("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
		}
		jobMu.Lock()
		job, ok := jobStore[id]
		jobMu.Unlock()
		if !ok {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "job not found"})
		}
		if job.cmd == nil || job.cmd.Process == nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "job not running"})
		}
		// attempt graceful stop, then kill if necessary
		if err := job.cmd.Process.Kill(); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to stop job"})
		}
		job.Status = JobStopped
		end := time.Now()
		job.EndedAt = &end
		return c.JSON(http.StatusOK, map[string]string{"status": "stopped"})
	})

	// GET /job/info?id=123 -> returns status + stdout/stderr (truncated)
	e.GET("/job/info", func(c echo.Context) error {
		id := c.QueryParam("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
		}
		jobMu.Lock()
		job, ok := jobStore[id]
		jobMu.Unlock()
		if !ok {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "job not found"})
		}

		stdout := truncate(job.Stdout.Bytes(), maxOutputReturn)
		stderr := truncate(job.Stderr.Bytes(), maxOutputReturn)

		resp := map[string]interface{}{
			"id":     job.ID,
			"cmd":    job.Cmd,
			"status": job.Status,
			"stdout": stdout,
			"stderr": stderr,
		}
		if job.ExitCode != nil {
			resp["exit_code"] = *job.ExitCode
		}
		if job.StartedAt != nil {
			resp["started_at"] = job.StartedAt.Format(time.RFC3339)
		}
		if job.EndedAt != nil {
			resp["ended_at"] = job.EndedAt.Format(time.RFC3339)
		}
		return c.JSON(http.StatusOK, resp)
	})

	// --- end job manager ---

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
