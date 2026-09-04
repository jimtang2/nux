set dotenv-load := true

test:
    go test ./lib/...

test-verbose:
    go test -v ./lib/...

test-watch:
    watchexec -r -e go -- just test

bench:
    go test -run=^$ -bench=. -benchmem ./lib/client/...

bench-watch:
    watchexec -r -e go -- just bench

run:
    go run cmd/*

run-watch:
    air

build:
    go build -o $GOBIN/nux cmd/*

build-watch:
    watchexec -r -e go -- just build