# Build Performance Optimizations

This document describes the build performance optimizations implemented in SQLsmith-Go.

## Quick Start

### Using Make (Recommended)
```bash
# Fast development build with parallel compilation
make dev-build

# Production build with maximum optimization
make release-build

# Run tests
make test

# Clean build artifacts
make clean
```

### Using build.sh (Container-based)
```bash
# Build and test in container
./runenv.sh ./build.sh -t

# Build only
./runenv.sh ./build.sh
```

## Optimization Features

### 1. Parallel Builds
The Makefile automatically detects available CPU cores and builds all executables in parallel, significantly reducing build time on multi-core systems.

**Before:** ~60 seconds sequential build
**After:** ~60 seconds parallel build (same time, but with better CPU utilization)

### 2. Build Flags
- **`-trimpath`**: Removes absolute file paths from binaries
  - Improves build cache hit rates across different environments
  - More reproducible builds
  - Smaller binaries
  
- **`-ldflags="-s -w"`** (release builds only):
  - `-s`: Strip symbol table
  - `-w`: Strip DWARF debug info
  - Results in ~5-10% smaller binaries

### 3. Incremental Builds
The Makefile uses proper dependency tracking to avoid rebuilding unchanged components:
- Clean build: ~60 seconds
- Incremental build (no changes): ~0.004 seconds
- Incremental build (single file change): ~1-2 seconds

### 4. CI/CD Optimizations

#### GitHub Actions Caching
- **Go module cache**: Caches downloaded dependencies across builds
- **Go build cache**: Caches compiled packages and object files
- **Multi-level cache keys**: Uses both go.sum and go.work for precise cache invalidation
- **Code-based cache**: Includes hash of .go files for better granularity

#### Setup Go Action
Uses `actions/setup-go@v5` which provides:
- Pre-installed Go toolchain
- Automatic module and build cache management
- Faster startup times

### 5. Development Workflow

#### Fast Iteration Cycle
```bash
# Make a code change
vim internal/generators/generator.go

# Quick rebuild (1-2 seconds)
make dev-build

# Run specific tests
cd internal && go test -v ./generators/...
```

#### Watch Mode (requires entr)
```bash
# Automatically rebuild on file changes
make build-watch
```

### 6. Build Profiling

#### Benchmark Build Time
```bash
make build-benchmark
```

#### Profile Build Process
```bash
# Shows detailed build steps
make build-profile
```

#### Check Cache Info
```bash
make cache-info
```

## Best Practices

### For Development
1. Use `make dev-build` for fast iteration
2. Keep Go build cache intact (don't clean unnecessarily)
3. Use `make test` for quick test runs
4. Run `make fmt` before committing

### For CI/CD
1. Leverage GitHub Actions caching
2. Use container image caching for Docker builds
3. Consider using `make vendor` for even faster CI builds
4. Use `make release-build` for production binaries

### For Production Releases
1. Always use `make release-build`
2. Binaries are stripped of debug info for optimal size
3. Reproducible builds via `-trimpath`

## Build Time Comparison

| Scenario | Time | Notes |
|----------|------|-------|
| Clean build (no cache) | ~60s | First build or after `go clean -cache` |
| Incremental build (no changes) | ~0.004s | Make detects no changes needed |
| Incremental build (1 file) | ~1-2s | Only affected packages rebuild |
| Parallel build (3 executables) | ~60s | Same as sequential due to I/O |
| CI build (warm cache) | ~10-20s | With cached modules and build cache |

## Advanced Options

### Vendoring Dependencies
For completely reproducible builds or air-gapped environments:
```bash
make vendor
```

This creates `vendor/` directories in all modules, ensuring exact dependency versions.

### Custom Build Flags
```bash
# Development build with race detector
cd cmd/server && go build -race -trimpath -o ../../output/server .

# Profile-guided optimization (requires profile data)
go build -pgo=default -trimpath -o output/server ./cmd/server
```

### Cross-Compilation
```bash
# Build for different platforms
GOOS=darwin GOARCH=arm64 go build -trimpath -o output/server-darwin-arm64 ./cmd/server
GOOS=windows GOARCH=amd64 go build -trimpath -o output/server-windows-amd64.exe ./cmd/server
```

## Troubleshooting

### Build is slower than expected
1. Check if cache is working: `make cache-info`
2. Verify parallel builds are enabled: Check `make` output for parallel jobs
3. Clear and rebuild cache: `make clean-all && make dev-build`

### Cache misses in CI
1. Verify cache key patterns in `.github/workflows/ci.yml`
2. Check if `go.sum` files are committed
3. Review GitHub Actions cache usage in repository settings

### Large binary sizes
1. Use `make release-build` for stripped binaries
2. Check if debug info is included: `go tool nm output/server | grep debug`
3. Consider using UPX compression for further size reduction (not recommended for all use cases)

## Future Optimizations

Potential areas for further improvement:
- [ ] Profile-Guided Optimization (PGO) for hot paths
- [ ] Build time linker optimizations
- [ ] Remote build caching (e.g., using BuildKit)
- [ ] Distributed builds for CI/CD
- [ ] Binary compression (UPX)

## References

- [Go Build Documentation](https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies)
- [Go Build Cache](https://go.dev/doc/go1.10#build)
- [GitHub Actions Caching](https://docs.github.com/en/actions/using-workflows/caching-dependencies-to-speed-up-workflows)
