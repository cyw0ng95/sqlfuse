package executors

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"sqlfuse/internal/common"
	"sqlfuse/internal/generators"
	"sqlfuse/internal/stmts/stmts"
)

// Executor is a simple command-line based executor abstraction.
// Implementations represent an executable binary and can build an *exec.Cmd for running it.
type Executor interface {
	Name() string
	Path() string
	BuildCmd(args []string, seed *int64) *exec.Cmd
}

// CmdExecutor implements Executor for a binary on disk.
type CmdExecutor struct {
	name string
	path string
}

func NewCmdExecutor(name, path string) *CmdExecutor {
	return &CmdExecutor{name: name, path: path}
}

func (c *CmdExecutor) Name() string { return c.name }
func (c *CmdExecutor) Path() string { return c.path }
func (c *CmdExecutor) BuildCmd(args []string, seed *int64) *exec.Cmd {
	// If seed is provided, add as a flag before other args.
	if seed != nil {
		seedArg := fmt.Sprintf("--seed=%d", *seed)
		args = append([]string{seedArg}, args...)
	}
	return exec.Command(c.path, args...)
}

// Run is the shared executor loop used by embedded executors. It centralizes
// logger initialization, DB connection, optional schema initialization, worker
// loop and token aggregation. The connect callback should open and return an
// *sql.DB for the provided DSN. genFactory constructs a generators.Generator
// for a worker seed. printSchema is invoked before and after execution to allow
// executors to emit discovered schema information.
func Run(startMsg string, flags *CommonFlags, connect func(dsn string) (*sql.DB, error), genFactory func(seed uint64) generators.Generator, printSchema func(db *sql.DB)) error {
	common.InitLogger()
	common.Logger.Info().Msg(startMsg)

	conn, err := connect(flags.Dsn)
	if err != nil {
		common.Logger.Error().Err(err).Msg("Error opening database")
		os.Exit(1)
	}
	defer conn.Close()

	// Initialize schema if provided
	if strings.TrimSpace(flags.InitSQLPath) != "" {
		initSQL, err := os.ReadFile(flags.InitSQLPath)
		if err != nil {
			common.Logger.Error().Err(err).Str("path", flags.InitSQLPath).Msg("Failed to read init SQL file")
			os.Exit(1)
		}
		common.Logger.Info().Msgf("Initializing database schema from %s", flags.InitSQLPath)
		for _, stmt := range strings.Split(string(initSQL), ";") {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := conn.Exec(stmt); err != nil {
				common.Logger.Error().Err(err).Str("stmt", stmt).Msg("Init SQL error")
				os.Exit(1)
			}
		}
	}

	printSchema(conn)

	workers := flags.Workers
	queries := flags.Queries
	if workers < 1 {
		workers = 1
	}
	if queries < 1 {
		queries = 1
	}

	// determine base seed: use provided seed if non-zero, otherwise time-based
	var baseSeed uint64
	if flags.Seed != 0 {
		baseSeed = uint64(flags.Seed)
	} else {
		baseSeed = uint64(time.Now().UnixNano())
	}

	// Parse custom weights if provided
	var customWeights map[stmts.StmtType]uint64
	if strings.TrimSpace(flags.Weights) != "" {
		var weightsMap map[string]uint64
		if err := json.Unmarshal([]byte(flags.Weights), &weightsMap); err != nil {
			common.Logger.Error().Err(err).Msg("Failed to parse weights JSON")
			os.Exit(1)
		}
		// Convert string keys to StmtType
		customWeights = make(map[stmts.StmtType]uint64, len(weightsMap))
		for k, v := range weightsMap {
			customWeights[stmts.StmtType(k)] = v
		}
		common.Logger.Info().Msgf("Using custom weights with %d statement types", len(customWeights))
	}

	var wg sync.WaitGroup
	wg.Add(workers)
	tokenCh := make(chan uint64, workers)

	for w := 0; w < workers; w++ {
		go func(workerID int) {
			defer wg.Done()
			// each worker gets a deterministic seed derived from base
			gen := genFactory(baseSeed + uint64(workerID))
			
			// Apply custom weights if provided
			if customWeights != nil {
				if bg, ok := gen.(*generators.TursoGenerator); ok {
					bg.SetWeights(customWeights)
				} else if bg, ok := gen.(*generators.GoSQLite3Generator); ok {
					bg.SetWeights(customWeights)
				} else if bg, ok := gen.(*generators.DuckDBGenerator); ok {
					bg.SetWeights(customWeights)
				}
			}
			
			for i := 0; i < queries; i++ {
				query := gen.GenerateWithDB(conn)
				_, execErr := conn.Exec(query)
				if execErr != nil {
					if flags.Verbose {
						common.Logger.Info().Msgf("Worker %d executing query %d: %s", workerID, i+1, query)
					}
					continue
				}
				if flags.Verbose {
					common.Logger.Info().Msgf("Worker %d executing query %d: %s", workerID, i+1, query)
				}
			}
			tokenCh <- gen.TokensUsed()
		}(w)
	}

	wg.Wait()
	close(tokenCh)

	var totalTokens uint64
	for t := range tokenCh {
		totalTokens += t
	}

	common.Logger.Info().Msgf("Total queries executed: %d", workers*queries)
	common.Logger.Info().Msgf("Total tokens used: %d", totalTokens)

	printSchema(conn)
	return nil
}

// ResolveExecPath ensures the requested executable refers to a file under outputDir
// and that the file exists and is executable. Returns the absolute path to the executable.
func ResolveExecPath(outputDir, requested string) (string, error) {
	if requested == "" {
		return "", errors.New("empty executable name")
	}

	var candidate string
	if filepath.IsAbs(requested) {
		candidate = requested
	} else {
		candidate = filepath.Join(outputDir, requested)
	}

	abs, err := filepath.Abs(candidate)
	if err != nil {
		return "", errors.New("invalid executable path")
	}

	// ensure it is under outputDir
	absOut, err := filepath.Abs(outputDir)
	if err != nil {
		return "", errors.New("invalid output directory")
	}
	absOutWithSep := absOut + string(os.PathSeparator)
	if !(abs == absOut || (len(abs) > len(absOutWithSep)-1 && (abs == absOut || filepath.HasPrefix(abs, absOutWithSep)))) {
		// fallback: ensure prefix match using filepath.Rel
		rel, rerr := filepath.Rel(absOut, abs)
		if rerr != nil || stringsHasDotDot(rel) {
			return "", errors.New("executable must reside inside output directory")
		}
	}

	st, err := os.Stat(abs)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", errors.New("executable not found in output directory")
		}
		return "", errors.New("failed to stat executable")
	}
	if st.IsDir() {
		return "", errors.New("executable path is a directory")
	}
	if st.Mode()&0111 == 0 {
		return "", errors.New("file is not executable")
	}
	return abs, nil
}

// DiscoverExecutors walks the output directory and returns executables discovered there.
func DiscoverExecutors(outputDir string) ([]Executor, error) {
	var out []Executor
	absOut, err := filepath.Abs(outputDir)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(absOut)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		p := filepath.Join(absOut, e.Name())
		st, err := os.Stat(p)
		if err != nil {
			continue
		}
		if st.Mode()&0111 == 0 {
			continue
		}
		out = append(out, NewCmdExecutor(e.Name(), p))
	}
	return out, nil
}

// small helper to detect .. in relative path without importing path/filepath in callers
func stringsHasDotDot(s string) bool {
	if s == "" {
		return false
	}
	for _, part := range filepath.SplitList(s) {
		if part == ".." {
			return true
		}
	}
	return false
}
