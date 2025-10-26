package executors

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
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
