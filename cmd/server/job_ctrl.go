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

	"sqlfuse/internal/executors"
)

// --- Job manager implementation ---
// Simple in-memory job store. Jobs invoke executables/commands under ./output and capture stdout/stderr.

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

// JobMeta is the persisted representation of a job (no buffers/cmd)
type JobMeta struct {
	ID        string     `json:"id"`
	Cmd       string     `json:"cmd"`
	Status    JobStatus  `json:"status"`
	PID       int        `json:"pid,omitempty"`
	ExitCode  *int       `json:"exit_code,omitempty"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

var (
	jobStore     = make(map[string]*Job)
	jobMu        sync.Mutex
	jobIDCounter uint64
)

// helper to truncate output for responses
func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "\n...(truncated)"
}

// saveJobs persists job metadata to disk (best-effort)
func saveJobs() {
	jobMu.Lock()
	defer jobMu.Unlock()
	// prepare directory
	dir := filepath.Dir(serverConfig.Job.PersistPath)
	_ = os.MkdirAll(dir, 0o755)

	metas := make([]JobMeta, 0, len(jobStore))
	for _, j := range jobStore {
		m := JobMeta{
			ID:        j.ID,
			Cmd:       j.Cmd,
			Status:    j.Status,
			PID:       j.PID,
			ExitCode:  j.ExitCode,
			StartedAt: j.StartedAt,
			EndedAt:   j.EndedAt,
		}
		metas = append(metas, m)
	}
	b, err := json.MarshalIndent(metas, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(serverConfig.Job.PersistPath, b, 0o644)
}

// loadJobs loads persisted jobs from disk into memory (best-effort)
func loadJobs() {
	b, err := os.ReadFile(serverConfig.Job.PersistPath)
	if err != nil {
		return
	}
	var metas []JobMeta
	if err := json.Unmarshal(b, &metas); err != nil {
		return
	}
	jobMu.Lock()
	defer jobMu.Unlock()
	for _, m := range metas {
		// only restore metadata; stdout/stderr/cmd remain empty
		jobStore[m.ID] = &Job{
			ID:        m.ID,
			Cmd:       m.Cmd,
			Status:    m.Status,
			PID:       m.PID,
			ExitCode:  m.ExitCode,
			StartedAt: m.StartedAt,
			EndedAt:   m.EndedAt,
		}
		// update jobIDCounter to avoid collisions
		if v, err := strconv.ParseUint(m.ID, 10, 64); err == nil {
			if v > jobIDCounter {
				jobIDCounter = v
			}
		}
	}
}

// RegisterJobRoutes registers job-related HTTP endpoints on the provided Echo instance.
func RegisterJobRoutes(e *echo.Echo) {
	// Prepare absolute output dir for validation
	outputDir, _ := filepath.Abs("./output")

	// load persisted jobs if any
	loadJobs()

	// POST /job/new accepts either:
	// { "executor": "turso_embedded", "args": ["--workers", "4"], "seed": 123, "weights": {"insert": 100} }
	// or legacy: { "cmd": "turso_embedded --workers 4", "seed": 123 }
	e.POST("/job/new", func(c echo.Context) error {
		var req struct {
			Cmd      string              `json:"cmd"`
			Executor string              `json:"executor"`
			Args     []string            `json:"args"`
			Seed     *int64              `json:"seed,omitempty"`
			Weights  map[string]uint64   `json:"weights,omitempty"`
		}
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}

		var execName string
		var args []string

		// Handle new format: executor + args
		if req.Executor != "" {
			execName = req.Executor
			args = req.Args
			if args == nil {
				args = []string{}
			}
		} else if strings.TrimSpace(req.Cmd) != "" {
			// Handle legacy format: cmd string
			toks := strings.Fields(req.Cmd)
			if len(toks) == 0 {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "empty command"})
			}
			execName = toks[0]
			if len(toks) > 1 {
				args = toks[1:]
			} else {
				args = []string{}
			}
		} else {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "either executor or cmd is required"})
		}

		// Look up executor in configured executors map
		var absCmdPath string
		if exec, ok := executorsByName[execName]; ok {
			// Use the configured executor path
			absCmdPath = exec.Path()
		} else {
			// Fallback: try to resolve executable path under ./output
			var err error
			absCmdPath, err = executors.ResolveExecPath(outputDir, execName)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("executor '%s' not found: %v", execName, err)})
			}
		}

		// build executor and exec.Cmd (forward optional seed)
		execBin := executors.NewCmdExecutor(filepath.Base(absCmdPath), absCmdPath)
		cmdObj := execBin.BuildCmd(args, req.Seed)

		// Ensure seed flag is present in cmd args (some callers may not honor prepended seed)
		if req.Seed != nil {
			seedArg := "--seed=" + strconv.FormatInt(*req.Seed, 10)
			found := false
			for _, a := range cmdObj.Args[1:] {
				if strings.HasPrefix(a, "--seed=") {
					found = true
					break
				}
			}
			if !found {
				cmdObj.Args = append(cmdObj.Args, seedArg)
			}
		}

		// Add weights flag if custom weights provided
		if req.Weights != nil && len(req.Weights) > 0 {
			weightsJSON, err := json.Marshal(req.Weights)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid weights format"})
			}
			weightsArg := "--weights=" + string(weightsJSON)
			cmdObj.Args = append(cmdObj.Args, weightsArg)
		}

		// debug: print final args to stderr
		if len(cmdObj.Args) > 0 {
			fmt.Fprintf(os.Stderr, "Starting job cmd args: %v\n", cmdObj.Args)
		}

		id := strconv.FormatUint(atomic.AddUint64(&jobIDCounter, 1), 10)
		fullCmd := absCmdPath
		if len(args) > 0 {
			fullCmd = absCmdPath + " " + strings.Join(args, " ")
		}
		if req.Seed != nil {
			fullCmd = fullCmd + " --seed=" + strconv.FormatInt(*req.Seed, 10)
		}
		job := &Job{
			ID:     id,
			Cmd:    fullCmd,
			Status: JobPending,
		}

		jobMu.Lock()
		jobStore[id] = job
		jobMu.Unlock()

		// persist
		saveJobs()

		// start the job asynchronously
		go func(j *Job, cmdToRun *exec.Cmd) {
			// mark running
			jobMu.Lock()
			j.Status = JobRunning
			now := time.Now()
			j.StartedAt = &now
			jobMu.Unlock()
			saveJobs()

			j.cmd = cmdToRun
			cmdToRun.Stdout = &j.Stdout
			cmdToRun.Stderr = &j.Stderr

			if err := cmdToRun.Start(); err != nil {
				jobMu.Lock()
				j.Status = JobFailed
				end := time.Now()
				j.EndedAt = &end
				jobMu.Unlock()
				saveJobs()
				return
			}

			if cmdToRun.Process != nil {
				jobMu.Lock()
				j.PID = cmdToRun.Process.Pid
				jobMu.Unlock()
				saveJobs()
			}

			// wait for completion
			err := cmdToRun.Wait()
			end := time.Now()
			jobMu.Lock()
			j.EndedAt = &end
			if err != nil {
				if cmdToRun.ProcessState != nil {
					exit := cmdToRun.ProcessState.ExitCode()
					j.ExitCode = &exit
				}
				if j.Status != JobStopped {
					j.Status = JobFailed
				}
				jobMu.Unlock()
				saveJobs()
				return
			}

			if cmdToRun.ProcessState != nil {
				exit := cmdToRun.ProcessState.ExitCode()
				j.ExitCode = &exit
			}
			j.Status = JobDone
			jobMu.Unlock()
			saveJobs()
		}(job, cmdObj)

		return c.JSON(http.StatusAccepted, map[string]string{"id": id})
	})

	// GET /job/status?id=123
	e.GET("/job/status", func(c echo.Context) error {
		id := c.QueryParam("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "id is required"})
		}
		// copy fields under lock to avoid races
		jobMu.Lock()
		job, ok := jobStore[id]
		if !ok {
			jobMu.Unlock()
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
		jobMu.Unlock()
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
		jobMu.Lock()
		job.Status = JobStopped
		end := time.Now()
		job.EndedAt = &end
		jobMu.Unlock()
		saveJobs()
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
		if !ok {
			jobMu.Unlock()
			return c.JSON(http.StatusNotFound, map[string]string{"error": "job not found"})
		}
		// copy buffers and metadata under lock
		stdoutBytes := append([]byte(nil), job.Stdout.Bytes()...)
		stderrBytes := append([]byte(nil), job.Stderr.Bytes()...)
		resp := map[string]interface{}{
			"id":     job.ID,
			"cmd":    job.Cmd,
			"status": job.Status,
			"stdout": truncate(stdoutBytes, serverConfig.Job.MaxOutputBytes),
			"stderr": truncate(stderrBytes, serverConfig.Job.MaxOutputBytes),
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
		jobMu.Unlock()
		return c.JSON(http.StatusOK, resp)
	})
}
