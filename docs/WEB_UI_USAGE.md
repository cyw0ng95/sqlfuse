# Web UI Usage Guide

## Overview

The SQLsmith-Go web UI provides an interface for managing fuzzing jobs across different database flavors. This guide explains how to use the interface to start and monitor fuzzing jobs.

## Starting the Server

1. Build the server:
   ```bash
   go build ./cmd/server
   ```

2. Configure executors in `config/executors.json`:
   ```json
   [
       {
           "executor": "turso_embedded",
           "path": "./output/turso_embedded",
           "flavor": "turso"
       },
       {
           "executor": "go_sqlite3_embedded",
           "path": "./output/go_sqlite3_embedded",
           "flavor": "go-sqlite3"
       }
   ]
   ```

3. Start the server:
   ```bash
   ./server
   ```

4. The server will start on port 8080 by default (configurable via `PORT` environment variable or `config/server.json`)

## Using the Web Interface

### Accessing the UI

1. Navigate to the view directory and start the development server:
   ```bash
   cd view
   pnpm install
   pnpm run dev
   ```

2. Open your browser to `http://localhost:3000`

### Starting a Fuzzing Job

The Jobs section allows you to start fuzzing jobs for different database flavors:

1. **Select an Executor**: Choose from the dropdown which shows both the executor name and its flavor:
   - `turso_embedded (turso)` - Turso LibSQL flavor
   - `go_sqlite3_embedded (go-sqlite3)` - Full SQLite3 support
   - `duckdb_embedded (duckdb)` - DuckDB flavor

2. **Specify Arguments**: Add command-line arguments for the executor, such as:
   - `--workers 4` - Number of concurrent workers
   - `--queries 100` - Number of queries per worker
   - `--init-sql /path/to/schema.sql` - Initialize database schema
   - `--verbose` - Show SQL queries being executed

3. **Start the Job**: Click the "Start Job" button to launch the fuzzing job

4. **Note the Job ID**: The UI will display the job ID once started

### Monitoring Jobs

After starting a job, you can monitor its progress:

1. **Check Status**: Enter the job ID and click "Status" to see:
   - Current status (pending, running, done, failed, stopped)
   - Process ID (PID) if running
   - Start and end times
   - Exit code

2. **View Output**: Click "Info" to see:
   - Complete job information
   - Stdout output (query execution logs)
   - Stderr output (error messages and warnings)

3. **Stop a Job**: Click "Stop" to terminate a running job

### Understanding Flavors

Each database flavor has different capabilities and restrictions:

**Turso LibSQL (`turso`)**:
- SQLite-compatible with some restrictions
- Does not support: window functions, recursive CTEs, REGEXP operator, etc.
- Ideal for testing Turso-specific deployments

**go-sqlite3 (`go-sqlite3`)**:
- Full SQLite3 feature support
- Supports: window functions, recursive CTEs, all operators, etc.
- Ideal for comprehensive SQLite3 testing

**DuckDB (`duckdb`)**:
- DuckDB database flavor
- Feature set depends on DuckDB implementation

## API Endpoints

The backend provides the following REST API endpoints:

- `GET /health` - Health check
- `GET /info` - Server information
- `GET /executors` - List available executors with flavor information
- `GET /generators/get` - Get generator metadata
- `POST /job/new` - Create a new job
  - Request body: `{"executor": "turso_embedded", "args": ["--workers", "4"], "seed": 42}`
- `GET /job/status?id=<job_id>` - Get job status
- `GET /job/info?id=<job_id>` - Get job output
- `POST /job/stop?id=<job_id>` - Stop a running job

## Example Workflow

1. Start a Turso fuzzing job:
   - Select `turso_embedded (turso)`
   - Arguments: `--workers 2 --queries 50 --init-sql /path/to/schema.sql`
   - Click "Start Job"

2. Start a go-sqlite3 fuzzing job in parallel:
   - Select `go_sqlite3_embedded (go-sqlite3)`
   - Arguments: `--workers 4 --queries 100 --init-sql /path/to/schema.sql`
   - Click "Start Job"

3. Monitor both jobs using their respective job IDs

4. Compare results between flavors to identify flavor-specific issues

## Troubleshooting

**Job fails immediately**:
- Check that executors are built and paths in `config/executors.json` are correct
- Verify init SQL file exists if specified
- Check stderr output for error messages

**Executor not found**:
- Ensure executors are built: `go build -o ./output/<executor_name> ./cmd/executors/<executor_dir>`
- Verify executor paths in `config/executors.json` point to valid executables

**Port already in use**:
- Change the server port via `PORT` environment variable: `PORT=8081 ./server`
- Or update `config/server.json`
