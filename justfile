set dotenv-load := true

test:
    go test ./lib/...

test-generator:
    go test ./lib/generator/...

test-verbose:
    go test -v ./lib/...

test-watch:
    watchexec -r -e go -- just test

bench:
    go test -run=^$ -bench=. -benchmem ./lib/client/...

bench-watch:
    watchexec -r -e go -- just bench

build:
    go build -o $GOBIN/nux cmd/nux/*

build-watch:
    watchexec -r -e go -- just build

run-collector:
    bin/otelcorecol_darwin_arm64 --config lib/testdata/otel-collector-config-kafka.yaml

run-telemetrygen:
    bin/telemetrygen logs --otlp-endpoint=localhost:4488 --otlp-http --otlp-http-url-path=/v1/logs --otlp-insecure --logs=10 --interval=10ms

run-kcat:
    bin/kcat -b localhost:9092 -X security.protocol=SASL_PLAINTEXT -X sasl.mechanisms=PLAIN -X sasl.username=user1 -X sasl.password=GjLy8ctKry -C -t nux-logs
