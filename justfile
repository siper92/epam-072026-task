image := "ticTacSolved-task-dev"

default:
    @just --list

build:
    docker build -t {{image}} -f _env/dev/Dockerfile _env

compile:
    docker run --rm -v {{justfile_directory()}}:/app -w /app/game/data {{image}} sqlc generate

test:
    go test ./...
