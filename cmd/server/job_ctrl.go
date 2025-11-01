package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	"sqlfuse/internal/common"
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
	Signal    *string    `json:"signal,omitempty"`    // Signal that killed the process, if any
	StartedAt *time.Time `json:"started_at,omitempty"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`

	Stdout bytes.Buffer `json:"-"`
	Stderr bytes.Buffer `json:"-"`

	cmd *exec.Cmd `json:"-"`
	
	// WebSocket support for streaming logs
	logSubscribers   map[*websocket.Conn]bool
	subscribersMutex sync.Mutex
}

// JobMeta is the persisted representation of a job (no buffers/cmd)
type JobMeta struct {
	ID        string     `json:"id"`
	Cmd       string     `json:"cmd"`
	Status    JobStatus  `json:"status"`
	PID       int        `json:"pid,omitempty"`
	ExitCode  *int       `json:"exit_code,omitempty"`
	Signal    *string    `json:"signal,omitempty"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

var (
	jobStore     = make(map[string]*Job)
	jobMu        sync.Mutex
	jobIDCounter uint64
	
	// WebSocket upgrader
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for development
		},
	}
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
			Signal:    j.Signal,
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
			ID:               m.ID,
			Cmd:              m.Cmd,
			Status:           m.Status,
			PID:              m.PID,
			ExitCode:         m.ExitCode,
			Signal:           m.Signal,
			StartedAt:        m.StartedAt,
			EndedAt:          m.EndedAt,
			logSubscribers:   make(map[*websocket.Conn]bool),
		}
		// update jobIDCounter to avoid collisions
		if v, err := strconv.ParseUint(m.ID, 10, 64); err == nil {
			if v > jobIDCounter {
				jobIDCounter = v
			}
		}
	}
}

// broadcastLogMessage sends a log message to all WebSocket subscribers of a job
func (j *Job) broadcastLogMessage(stream string, data string) {
	j.subscribersMutex.Lock()
	defer j.subscribersMutex.Unlock()
	
	msg := map[string]string{
		"stream": stream,
		"data":   data,
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		// This should rarely happen with simple string maps, but handle it
		return
	}
	
	for conn := range j.logSubscribers {
		if err := conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
			// Remove failed connection
			conn.Close()
			delete(j.logSubscribers, conn)
		}
	}
}

// addLogSubscriber adds a WebSocket connection to receive log updates
func (j *Job) addLogSubscriber(conn *websocket.Conn) {
	j.subscribersMutex.Lock()
	defer j.subscribersMutex.Unlock()
	j.logSubscribers[conn] = true
}

// removeLogSubscriber removes a WebSocket connection from log updates
func (j *Job) removeLogSubscriber(conn *websocket.Conn) {
	j.subscribersMutex.Lock()
	defer j.subscribersMutex.Unlock()
	delete(j.logSubscribers, conn)
}

// streamReader reads from a reader and broadcasts lines to WebSocket subscribers
func streamReader(j *Job, stream string, reader io.Reader, buffer *bytes.Buffer) {
	br := bufio.NewReader(reader)
	for {
		line, err := br.ReadString('\n')
		if len(line) > 0 {
			// Write to buffer
			buffer.WriteString(line)
			// Broadcast to WebSocket subscribers
			j.broadcastLogMessage(stream, line)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			// On read error, break to avoid tight loop
			break
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
				return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("invalid weights format: %v", err)})
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
			ID:             id,
			Cmd:            fullCmd,
			Status:         JobPending,
			logSubscribers: make(map[*websocket.Conn]bool),
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
			
			// Create pipes for stdout and stderr to enable streaming
			stdoutPipe, err := cmdToRun.StdoutPipe()
			if err != nil {
				jobMu.Lock()
				j.Status = JobFailed
				end := time.Now()
				j.EndedAt = &end
				jobMu.Unlock()
				saveJobs()
				return
			}
			
			stderrPipe, err := cmdToRun.StderrPipe()
			if err != nil {
				jobMu.Lock()
				j.Status = JobFailed
				end := time.Now()
				j.EndedAt = &end
				jobMu.Unlock()
				saveJobs()
				return
			}

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

			// Stream stdout and stderr concurrently
			var wg sync.WaitGroup
			wg.Add(2)
			
			go func() {
				defer wg.Done()
				streamReader(j, "stdout", stdoutPipe, &j.Stdout)
			}()
			
			go func() {
				defer wg.Done()
				streamReader(j, "stderr", stderrPipe, &j.Stderr)
			}()

			// wait for completion
			err = cmdToRun.Wait()
			
			// Wait for all output to be processed
			wg.Wait()
			
			end := time.Now()
			jobMu.Lock()
			j.EndedAt = &end
			if err != nil {
				if cmdToRun.ProcessState != nil {
					exit := cmdToRun.ProcessState.ExitCode()
					j.ExitCode = &exit
					
					// Check if process was killed by a signal
					if ws, ok := cmdToRun.ProcessState.Sys().(syscall.WaitStatus); ok {
						if ws.Signaled() {
							sig := ws.Signal().String()
							j.Signal = &sig
						}
					}
				}
				if j.Status != JobStopped {
					j.Status = JobFailed
				}
				jobMu.Unlock()
				saveJobs()
				
				// Send completion message to subscribers with signal info
				failureMsg := fmt.Sprintf("Job failed with exit code: %v", j.ExitCode)
				if j.Signal != nil {
					failureMsg = fmt.Sprintf("Job failed - killed by signal: %s (exit code: %v)", *j.Signal, j.ExitCode)
				}
				j.broadcastLogMessage("status", failureMsg)
				return
			}

			if cmdToRun.ProcessState != nil {
				exit := cmdToRun.ProcessState.ExitCode()
				j.ExitCode = &exit
			}
			j.Status = JobDone
			jobMu.Unlock()
			saveJobs()
			
			// Send completion message to subscribers
			j.broadcastLogMessage("status", "Job completed successfully")
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
		if job.Signal != nil {
			resp["signal"] = *job.Signal
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
		if job.Signal != nil {
			resp["signal"] = *job.Signal
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

	// WebSocket endpoint for streaming job logs
	// GET /job/logs/stream?id=123
	e.GET("/job/logs/stream", func(c echo.Context) error {
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

		// Upgrade HTTP connection to WebSocket
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}
		defer ws.Close()

		// Add this connection as a subscriber
		job.addLogSubscriber(ws)
		defer job.removeLogSubscriber(ws)

		// Send initial status message
		statusMsg := map[string]string{
			"stream": "status",
			"data":   fmt.Sprintf("Connected to job %s (status: %s)", id, job.Status),
		}
		if msgBytes, err := json.Marshal(statusMsg); err == nil {
			ws.WriteMessage(websocket.TextMessage, msgBytes)
		}

		// Send existing logs if any
		jobMu.Lock()
		// Create readers for the buffers to stream historical logs without large allocations.
		stdoutReader := bytes.NewReader(job.Stdout.Bytes())
		stderrReader := bytes.NewReader(job.Stderr.Bytes())
		jobMu.Unlock()

		// Stream historical stdout line-by-line
		stdoutScanner := bufio.NewScanner(stdoutReader)
		for stdoutScanner.Scan() {
			historyMsg := map[string]string{
				"stream": "stdout",
				"data":   stdoutScanner.Text() + "\n",
			}
			if msgBytes, err := json.Marshal(historyMsg); err == nil {
				if err := ws.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
					break // Stop sending if client disconnects
				}
			}
		}

		// Stream historical stderr line-by-line
		stderrScanner := bufio.NewScanner(stderrReader)
		for stderrScanner.Scan() {
			historyMsg := map[string]string{
				"stream": "stderr",
				"data":   stderrScanner.Text() + "\n",
			}
			if msgBytes, err := json.Marshal(historyMsg); err == nil {
				if err := ws.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
					break // Stop sending if client disconnects
				}
			}
		}

		// Keep connection alive and handle ping/pong
		ws.SetPongHandler(func(string) error {
			return nil
		})
		
		for {
			msgType, msg, err := ws.ReadMessage()
			if err != nil {
				break
			}
			// Handle ping messages by responding with pong
			if msgType == websocket.PingMessage {
				if err := ws.WriteMessage(websocket.PongMessage, nil); err != nil {
					break
				}
			}
			// Log unexpected messages for debugging
			if msgType == websocket.TextMessage || msgType == websocket.BinaryMessage {
				common.Logger.Debug().Msgf("Received unexpected message from client: %s", string(msg))
			}
		}

		return nil
	})
}
