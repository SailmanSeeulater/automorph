.PHONY: dev web gen-api build build-linux-arm64 build-linux-amd64 build-windows tidy db sqlc admin test

# Run these two in separate terminals, then open http://localhost:5180
dev:
	go run ./cmd/server

web:
	npm --prefix web run dev

# Regenerate web/src/api/schema.d.ts from the running server (make dev first)
gen-api:
	npm --prefix web run gen:api

build:
	go build -o bin/automorph ./cmd/server

# For the Oracle VPS. Run `uname -m` there: aarch64 = arm64, x86_64 = amd64.
build-linux-arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bin/automorph-linux-arm64 ./cmd/server

build-linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/automorph-linux-amd64 ./cmd/server

build-windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o bin/automorph.exe ./cmd/server

tidy:
	go mod tidy

db:
	sqlite3 data/automorph.db

# Regenerate internal/store from internal/db/queries
sqlc:
	sqlc generate

# Create or reset the admin: make admin EMAIL=you@example.com
admin:
	go run ./cmd/server create-admin -email $(EMAIL)

test:
	go test ./...