module sqlfuse/cmd/executors/duckdb_embedded

go 1.24.9

require (
	github.com/marcboeker/go-duckdb v1.8.3
	github.com/spf13/cobra v1.10.1
	sqlfuse/internal v0.0.0-00010101000000-000000000000
)

require (
	github.com/apache/arrow-go/v18 v18.0.0 // indirect
	github.com/goccy/go-json v0.10.4 // indirect
	github.com/google/flatbuffers v24.12.23+incompatible // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	golang.org/x/exp v0.0.0-20250813145105-42675adae3e6 // indirect
	golang.org/x/mod v0.27.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/tools v0.36.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
)

replace sqlfuse/internal => ../../../internal
