.PHONY: dev build build-linux-arm64 build-linux-amd64 build-windows tidy

dev:
	go run ./cmd/server

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