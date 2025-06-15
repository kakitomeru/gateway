SHELL := /bin/bash

start:
	(set -a; source .env; set +a; go run cmd/main.go)

update-pkg:
	go get -u github.com/kakitomeru/shared@main
	go get -u github.com/kakitomeru/auth@main
	go get -u github.com/kakitomeru/snippet@main
	go mod tidy