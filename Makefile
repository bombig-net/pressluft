SHELL := /bin/sh
.DEFAULT_GOAL := cli

# Bootstrap: build the pressluft CLI. After this, use 'pressluft <command>' for everything.
cli:
	go build -o bin/pressluft ./cmd/pressluft

clean:
	rm -rf bin/

.PHONY: cli clean
