image := "epam-task-dev"

default:
    @just --list

build:
    docker build -t epam_go_env -f _env/dev/Dockerfile _env

compile:
    docker run --rm -v {{justfile_directory()}}:/app -w /app/game/data epam_go_env sqlc generate

test:
    go test ./...
