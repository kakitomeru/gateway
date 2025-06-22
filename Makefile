SHELL := /bin/bash

start:
	(set -a; source .env; set +a; go run cmd/main.go)

update-pkg:
	GOPROXY=direct go get -u github.com/kakitomeru/shared@main
	GOPROXY=direct go get -u github.com/kakitomeru/auth@main
	GOPROXY=direct go get -u github.com/kakitomeru/snippet@main
	go mod tidy