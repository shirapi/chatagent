#!/bin/zsh

docker compose exec -w /workspace/src app goimports -w .
docker compose exec -w /workspace/src app staticcheck ./...
