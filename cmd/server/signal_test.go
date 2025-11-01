package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// TestSignalCapture tests that signal information is captured when a process is killed
func TestSignalCapture(t *testing.T) {
	// Initialize server config for testing
	serverConfig = ServerConfig{
		Port:                "8080",
		ExecutorsConfigPath: "/tmp/test-executors.json",
		ServerName:          "test server",
		ServerVersion:       "0.1",
		Job: JobConfig{
			MaxOutputBytes: 65536,
			PersistPath:    "/tmp/test-jobs-signal.json",
		},
	}

	// Create Echo instance
	e := echo.New()
	RegisterJobRoutes(e)

	// Start test server
	server := httptest.NewServer(e)
	defer server.Close()

	// Create a job that will be killed with SIGKILL
	cmd := exec.Command("bash", "-c", "sleep 60")
	
	jobID := "signal-test-1"
	job := &Job{
		ID:             jobID,
		Cmd:            "sleep 60",
		Status:         JobPending,
		logSubscribers: make(map[*websocket.Conn]bool),
	}

	jobMu.Lock()
	jobStore[jobID] = job
	jobMu.Unlock()

	// Start the job in a goroutine (similar to real job execution)
	done := make(chan bool)
	go func() {
		job.Status = JobRunning
		job.cmd = cmd
		
		if err := cmd.Start(); err != nil {
			t.Errorf("Failed to start command: %v", err)
			done <- true
			return
		}
		
		job.PID = cmd.Process.Pid
		
		// Wait for command to complete
		err := cmd.Wait()
		
		jobMu.Lock()
		if err != nil {
			if cmd.ProcessState != nil {
				exit := cmd.ProcessState.ExitCode()
				job.ExitCode = &exit
				
				// Check if process was killed by a signal
				if ws, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok {
					if ws.Signaled() {
						sig := ws.Signal().String()
						job.Signal = &sig
					}
				}
			}
			job.Status = JobFailed
		} else {
			job.Status = JobDone
		}
		jobMu.Unlock()
		done <- true
	}()

	// Wait for process to start
	time.Sleep(100 * time.Millisecond)

	// Kill the process with SIGKILL
	if job.cmd != nil && job.cmd.Process != nil {
		if err := job.cmd.Process.Signal(syscall.SIGKILL); err != nil {
			t.Fatalf("Failed to send SIGKILL: %v", err)
		}
	}

	// Wait for job to finish
	select {
	case <-done:
		// Job finished
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for job to finish")
	}

	// Verify that signal was captured
	jobMu.Lock()
	if job.Signal == nil {
		t.Error("Expected signal to be captured, but it was nil")
	} else if !strings.Contains(*job.Signal, "killed") && !strings.Contains(*job.Signal, "SIGKILL") {
		t.Errorf("Expected signal to contain 'killed' or 'SIGKILL', got: %s", *job.Signal)
	}
	
	if job.ExitCode == nil {
		t.Error("Expected exit code to be set")
	} else if *job.ExitCode != -1 {
		t.Errorf("Expected exit code -1 for killed process, got: %d", *job.ExitCode)
	}
	
	if job.Status != JobFailed {
		t.Errorf("Expected status JobFailed, got: %s", job.Status)
	}
	jobMu.Unlock()

	// Test that signal is included in status API response
	statusURL := server.URL + "/job/status?id=" + jobID
	resp, err := http.Get(statusURL)
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}
	defer resp.Body.Close()

	var statusResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		t.Fatalf("Failed to decode status response: %v", err)
	}

	if _, ok := statusResp["signal"]; !ok {
		t.Error("Expected 'signal' field in status response")
	}

	// Clean up
	jobMu.Lock()
	delete(jobStore, jobID)
	jobMu.Unlock()
}

// TestNormalExitNoSignal tests that signal is not set for normal process exit
func TestNormalExitNoSignal(t *testing.T) {
	// Initialize server config for testing
	serverConfig = ServerConfig{
		Port:                "8080",
		ExecutorsConfigPath: "/tmp/test-executors.json",
		ServerName:          "test server",
		ServerVersion:       "0.1",
		Job: JobConfig{
			MaxOutputBytes: 65536,
			PersistPath:    "/tmp/test-jobs-normal.json",
		},
	}

	// Create a job that exits normally
	cmd := exec.Command("bash", "-c", "exit 0")
	
	jobID := "normal-exit-test-1"
	job := &Job{
		ID:             jobID,
		Cmd:            "exit 0",
		Status:         JobPending,
		logSubscribers: make(map[*websocket.Conn]bool),
	}

	jobMu.Lock()
	jobStore[jobID] = job
	jobMu.Unlock()

	// Run the command
	job.Status = JobRunning
	job.cmd = cmd
	
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start command: %v", err)
	}
	
	job.PID = cmd.Process.Pid
	
	// Wait for command to complete
	err := cmd.Wait()
	
	jobMu.Lock()
	if err != nil {
		if cmd.ProcessState != nil {
			exit := cmd.ProcessState.ExitCode()
			job.ExitCode = &exit
			
			// Check if process was killed by a signal
			if ws, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok {
				if ws.Signaled() {
					sig := ws.Signal().String()
					job.Signal = &sig
				}
			}
		}
		job.Status = JobFailed
	} else {
		if cmd.ProcessState != nil {
			exit := cmd.ProcessState.ExitCode()
			job.ExitCode = &exit
		}
		job.Status = JobDone
	}
	jobMu.Unlock()

	// Verify that signal was NOT set for normal exit
	jobMu.Lock()
	if job.Signal != nil {
		t.Errorf("Expected signal to be nil for normal exit, got: %s", *job.Signal)
	}
	
	if job.ExitCode == nil {
		t.Error("Expected exit code to be set")
	} else if *job.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got: %d", *job.ExitCode)
	}
	
	if job.Status != JobDone {
		t.Errorf("Expected status JobDone, got: %s", job.Status)
	}
	jobMu.Unlock()

	// Clean up
	jobMu.Lock()
	delete(jobStore, jobID)
	jobMu.Unlock()
}

// TestSignalInJobMeta tests that signal information is persisted in JobMeta
func TestSignalInJobMeta(t *testing.T) {
	serverConfig = ServerConfig{
		Port:                "8080",
		ExecutorsConfigPath: "/tmp/test-executors.json",
		ServerName:          "test server",
		ServerVersion:       "0.1",
		Job: JobConfig{
			MaxOutputBytes: 65536,
			PersistPath:    "/tmp/test-jobs-meta.json",
		},
	}

	jobID := "meta-test-1"
	signal := "killed"
	exitCode := -1
	
	job := &Job{
		ID:             jobID,
		Cmd:            "test command",
		Status:         JobFailed,
		Signal:         &signal,
		ExitCode:       &exitCode,
		logSubscribers: make(map[*websocket.Conn]bool),
	}

	jobMu.Lock()
	jobStore[jobID] = job
	jobMu.Unlock()

	// Save jobs
	saveJobs()

	// Load jobs
	jobMu.Lock()
	delete(jobStore, jobID)
	jobMu.Unlock()

	loadJobs()

	// Verify signal was persisted and loaded
	jobMu.Lock()
	loadedJob, ok := jobStore[jobID]
	jobMu.Unlock()

	if !ok {
		t.Fatal("Job not found after loading")
	}

	if loadedJob.Signal == nil {
		t.Error("Expected signal to be loaded, but it was nil")
	} else if *loadedJob.Signal != signal {
		t.Errorf("Expected signal %q, got %q", signal, *loadedJob.Signal)
	}

	// Clean up
	jobMu.Lock()
	delete(jobStore, jobID)
	jobMu.Unlock()
}
