.PHONY: local release

local:
	go build -o bin/ssgo cmd/ssgo/main.go

release:
	GOOS=linux GOARCH=amd64 go build -o bin/ssgo cmd/ssgo/main.go
