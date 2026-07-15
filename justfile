image := "ttt-dev"
build-image := "ttt-build-image"
prod-image := "ttt-server-image"
test-image := "ttt-test-image"

default:
    @just --list

build:
    docker build -t {{image}} -f _env/dev/Dockerfile _env

compile:
    docker run --rm -v {{justfile_directory()}}:/app -w /app/game/data {{image}} sqlc generate

build-test:
    docker build -t {{test-image}} -f _env/test/Dockerfile _env

test: build-test
    docker run --rm \
        -v {{justfile_directory()}}:/app \
        -v ttt-test-go-cache:/go \
        -e JWT_SECRET=${JWT_SECRET:-test-secret} \
        {{test-image}}

test-unit:
    go test ./...

run:
    JWT_SECRET=${JWT_SECRET:-dev-secret} go run ./cli/server/main serve

compile-check:
    #!/usr/bin/env bash
    set -euo pipefail
    checksum() {
        find game/data/gen -type f -exec shasum -a 256 {} + | sort | shasum -a 256
    }
    before=$(checksum)
    just compile
    after=$(checksum)
    if [ "$before" != "$after" ]; then
        echo "ABORT: compile regenerated code in game/data/gen"
        echo "review and commit the regenerated files, then build again"
        exit 2
    fi

    echo "compile check passed: generated code is up to date"

build-prod: compile-check
    docker build --target ttt-build-image -t {{build-image}} -f _env/prod/Dockerfile .
    docker build --target ttt-server-image -t {{prod-image}} -f _env/prod/Dockerfile .

run-prod:
    docker run --rm -e JWT_SECRET=${JWT_SECRET:-change-me} -p 8080:8080 {{prod-image}}
