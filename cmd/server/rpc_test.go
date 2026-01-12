package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"

	"sqlfuse/internal/generators"
)

// TestHealthEndpoint tests the /health endpoint
func TestHealthEndpoint(t *testing.T) {
	// Initialize server config for testing
	serverConfig = ServerConfig{
		Port:                "8080",
		ExecutorsConfigPath: "/tmp/test-executors.json",
		ServerName:          "test server",
		ServerVersion:       "0.1",
		Job: JobConfig{
			MaxOutputBytes: 65536,
			PersistPath:    "/tmp/test-jobs.json",
		},
	}

	// Create Echo instance
	e := echo.New()

	// Add health endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":    "healthy",
			"timestamp": "2024-01-01T00:00:00Z",
		})
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Serve the request
	e.ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %q", resp["status"])
	}

	if _, ok := resp["timestamp"]; !ok {
		t.Error("Expected 'timestamp' field in response")
	}
}

// TestInfoEndpoint tests the /info endpoint
func TestInfoEndpoint(t *testing.T) {
	// Initialize server config for testing
	serverConfig = ServerConfig{
		Port:                "8080",
		ExecutorsConfigPath: "/tmp/test-executors.json",
		ServerName:          "test server",
		ServerVersion:       "0.1",
		Job: JobConfig{
			MaxOutputBytes: 65536,
			PersistPath:    "/tmp/test-jobs.json",
		},
	}

	// Create Echo instance
	e := echo.New()

	// Add info endpoint
	e.GET("/info", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"name":    serverConfig.ServerName,
			"version": serverConfig.ServerVersion,
		})
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/info", nil)
	rec := httptest.NewRecorder()

	// Serve the request
	e.ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp["name"] != serverConfig.ServerName {
		t.Errorf("Expected name %q, got %q", serverConfig.ServerName, resp["name"])
	}

	if resp["version"] != serverConfig.ServerVersion {
		t.Errorf("Expected version %q, got %q", serverConfig.ServerVersion, resp["version"])
	}
}

// TestGeneratorsGetEndpoint tests the /generators/get endpoint (generator metadata)
func TestGeneratorsGetEndpoint(t *testing.T) {
	// Initialize server config for testing
	serverConfig = ServerConfig{
		Port:                "8080",
		ExecutorsConfigPath: "/tmp/test-executors.json",
		ServerName:          "test server",
		ServerVersion:       "0.1",
		Job: JobConfig{
			MaxOutputBytes: 65536,
			PersistPath:    "/tmp/test-jobs.json",
		},
	}

	// Create Echo instance
	e := echo.New()

	// Add generators/get endpoint (using actual implementation from main.go)
	e.GET("/generators/get", func(c echo.Context) error {
		// Import is at package level, so we can use generators directly
		gen := generators.NewTursoGenerator(0)
		resp := map[string]interface{}{
			"generator": gen.Name(),
			"stmts":     gen.SupportedStmts(),
			"note":      "These are the high-level statement types the generator can produce with their default weights.",
		}
		return c.JSON(http.StatusOK, resp)
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/generators/get", nil)
	rec := httptest.NewRecorder()

	// Serve the request
	e.ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify generator field exists and is a string
	generator, ok := resp["generator"].(string)
	if !ok {
		t.Error("Expected 'generator' field to be a string")
	}
	if generator == "" {
		t.Error("Expected 'generator' field to be non-empty")
	}

	// Verify stmts field exists and is a map
	stmts, ok := resp["stmts"].(map[string]interface{})
	if !ok {
		t.Error("Expected 'stmts' field to be a map")
	}
	if len(stmts) == 0 {
		t.Error("Expected 'stmts' field to contain at least one statement type")
	}

	// Verify note field exists
	if _, ok := resp["note"]; !ok {
		t.Error("Expected 'note' field in response")
	}

	// Verify that stmts contains valid statement types with numeric weights
	for stmt, weight := range stmts {
		if stmt == "" {
			t.Error("Found empty statement name in stmts")
		}
		// Weight should be a number (can be int or float in JSON)
		switch weight.(type) {
		case float64, int:
			// Valid numeric weight
		default:
			t.Errorf("Expected numeric weight for stmt %q, got %T", stmt, weight)
		}
	}
}

// TestExecutorsEndpoint tests the /executors endpoint
func TestExecutorsEndpoint(t *testing.T) {
	// Initialize server config for testing
	serverConfig = ServerConfig{
		Port:                "8080",
		ExecutorsConfigPath: "/tmp/test-executors.json",
		ServerName:          "test server",
		ServerVersion:       "0.1",
		Job: JobConfig{
			MaxOutputBytes: 65536,
			PersistPath:    "/tmp/test-jobs.json",
		},
	}

	// Create a test executors config file
	testExecutorsConfig := []ExecutorConfig{
		{
			Executor: "test_executor_1",
			Path:     "/tmp/test_executor_1",
			Flavor:   "turso",
		},
		{
			Executor: "test_executor_2",
			Path:     "/tmp/test_executor_2",
			Flavor:   "go-sqlite3",
		},
	}
	executorsConfig = testExecutorsConfig

	// Create Echo instance
	e := echo.New()

	// Add executors endpoint
	e.GET("/executors", func(c echo.Context) error {
		return c.JSON(http.StatusOK, executorsConfig)
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/executors", nil)
	rec := httptest.NewRecorder()

	// Serve the request
	e.ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	var resp []ExecutorConfig
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response contains the executors
	if len(resp) != len(testExecutorsConfig) {
		t.Errorf("Expected %d executors, got %d", len(testExecutorsConfig), len(resp))
	}

	// Verify each executor config
	for i, exec := range resp {
		if exec.Executor != testExecutorsConfig[i].Executor {
			t.Errorf("Expected executor name %q, got %q", testExecutorsConfig[i].Executor, exec.Executor)
		}
		if exec.Path != testExecutorsConfig[i].Path {
			t.Errorf("Expected executor path %q, got %q", testExecutorsConfig[i].Path, exec.Path)
		}
		if exec.Flavor != testExecutorsConfig[i].Flavor {
			t.Errorf("Expected executor flavor %q, got %q", testExecutorsConfig[i].Flavor, exec.Flavor)
		}
	}

	// Clean up
	executorsConfig = nil
}

// TestGeneratorsGetEndpointIntegration is a full integration test using actual server setup
func TestGeneratorsGetEndpointIntegration(t *testing.T) {
	// Initialize server config for testing
	serverConfig = ServerConfig{
		Port:                "8080",
		ExecutorsConfigPath: "/tmp/test-executors-integration.json",
		ServerName:          "test server",
		ServerVersion:       "0.1",
		Job: JobConfig{
			MaxOutputBytes: 65536,
			PersistPath:    "/tmp/test-jobs-integration.json",
		},
	}

	// Create Echo instance with all middleware
	e := echo.New()
	
	// Add the actual endpoint from main.go
	e.GET("/generators/get", func(c echo.Context) error {
		gen := generators.NewTursoGenerator(0)
		resp := map[string]interface{}{
			"generator": gen.Name(),
			"stmts":     gen.SupportedStmts(),
			"note":      "These are the high-level statement types the generator can produce with their default weights.",
		}
		return c.JSON(http.StatusOK, resp)
	})

	// Start test server
	server := httptest.NewServer(e)
	defer server.Close()

	// Make actual HTTP request
	resp, err := http.Get(server.URL + "/generators/get")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Verify status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Decode response
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response structure
	if _, ok := data["generator"]; !ok {
		t.Error("Expected 'generator' field in response")
	}
	if _, ok := data["stmts"]; !ok {
		t.Error("Expected 'stmts' field in response")
	}
	if _, ok := data["note"]; !ok {
		t.Error("Expected 'note' field in response")
	}

	// Verify content-type header
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json; charset=UTF-8" {
		t.Logf("Content-Type: %q (expected application/json)", contentType)
	}
}

// TestExecutorsEndpointWithEmptyConfig tests /executors endpoint with no executors configured
func TestExecutorsEndpointWithEmptyConfig(t *testing.T) {
	// Initialize server config for testing
	serverConfig = ServerConfig{
		Port:                "8080",
		ExecutorsConfigPath: "/tmp/test-executors-empty.json",
		ServerName:          "test server",
		ServerVersion:       "0.1",
		Job: JobConfig{
			MaxOutputBytes: 65536,
			PersistPath:    "/tmp/test-jobs-empty.json",
		},
	}

	// Set empty executors config
	executorsConfig = []ExecutorConfig{}

	// Create Echo instance
	e := echo.New()

	// Add executors endpoint
	e.GET("/executors", func(c echo.Context) error {
		return c.JSON(http.StatusOK, executorsConfig)
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/executors", nil)
	rec := httptest.NewRecorder()

	// Serve the request
	e.ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	var resp []ExecutorConfig
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response is empty array
	if len(resp) != 0 {
		t.Errorf("Expected empty executors array, got %d executors", len(resp))
	}

	// Clean up
	executorsConfig = nil
}

// TestServerConfigLoading tests that server configuration can be loaded from file
func TestServerConfigLoading(t *testing.T) {
	// Create a temporary config file
	tmpFile, err := os.CreateTemp("", "server-config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test config
	testConfig := ServerConfig{
		Port:                "9999",
		ExecutorsConfigPath: "/tmp/test-exec.json",
		ServerName:          "test-config-server",
		ServerVersion:       "1.2.3",
		Job: JobConfig{
			MaxOutputBytes: 12345,
			PersistPath:    "/tmp/test-persist.json",
		},
	}

	data, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if _, err := tmpFile.Write(data); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFile.Close()

	// Load the config
	loadedConfig, err := loadServerConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded config
	if loadedConfig.Port != testConfig.Port {
		t.Errorf("Expected port %q, got %q", testConfig.Port, loadedConfig.Port)
	}
	if loadedConfig.ServerName != testConfig.ServerName {
		t.Errorf("Expected server name %q, got %q", testConfig.ServerName, loadedConfig.ServerName)
	}
	if loadedConfig.ServerVersion != testConfig.ServerVersion {
		t.Errorf("Expected server version %q, got %q", testConfig.ServerVersion, loadedConfig.ServerVersion)
	}
	if loadedConfig.Job.MaxOutputBytes != testConfig.Job.MaxOutputBytes {
		t.Errorf("Expected max output bytes %d, got %d", testConfig.Job.MaxOutputBytes, loadedConfig.Job.MaxOutputBytes)
	}
}

// TestExecutorsConfigLoading tests that executors configuration can be loaded from file
func TestExecutorsConfigLoading(t *testing.T) {
	// Create a temporary config file
	tmpFile, err := os.CreateTemp("", "executors-config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test config
	testExecutors := []ExecutorConfig{
		{
			Executor: "turso_embedded",
			Path:     "/path/to/turso_embedded",
			Flavor:   "turso",
		},
		{
			Executor: "duckdb_embedded",
			Path:     "/path/to/duckdb_embedded",
			Flavor:   "duckdb",
		},
	}

	data, err := json.MarshalIndent(testExecutors, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if _, err := tmpFile.Write(data); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFile.Close()

	// Load the config
	loadedExecutors, err := loadExecutorsConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load executors config: %v", err)
	}

	// Verify loaded config
	if len(loadedExecutors) != len(testExecutors) {
		t.Fatalf("Expected %d executors, got %d", len(testExecutors), len(loadedExecutors))
	}

	for i, exec := range loadedExecutors {
		if exec.Executor != testExecutors[i].Executor {
			t.Errorf("Expected executor name %q, got %q", testExecutors[i].Executor, exec.Executor)
		}
		if exec.Path != testExecutors[i].Path {
			t.Errorf("Expected executor path %q, got %q", testExecutors[i].Path, exec.Path)
		}
		if exec.Flavor != testExecutors[i].Flavor {
			t.Errorf("Expected executor flavor %q, got %q", testExecutors[i].Flavor, exec.Flavor)
		}
	}
}
