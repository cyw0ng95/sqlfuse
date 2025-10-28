module sqlsmith-go/cmd/executors/go_sqlite3_embedded

go 1.24.9

require (
	github.com/mattn/go-sqlite3 v1.14.32
	github.com/spf13/cobra v1.10.1
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
)

replace sqlsmith-go/internal => ../../../internal
