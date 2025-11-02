module sqlfuse/cmd/executors/validator_turso_sqlite3

go 1.24.9

require (
	github.com/mattn/go-sqlite3 v1.14.32
	github.com/spf13/cobra v1.10.1
	github.com/tursodatabase/turso-go v0.2.2
	sqlfuse/internal v0.0.0
)

require (
	github.com/ebitengine/purego v0.10.0-alpha.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	golang.org/x/sys v0.37.0 // indirect
)

replace sqlfuse/internal => ../../../internal
