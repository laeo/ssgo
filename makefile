.PHONY: ssgo

ssgo:
	GOOS=linux GOARCH=amd64 go build -o bin/ssgo cmd/ssgo/main.go
