.PHONY: all clean

all: build
	go build -o bfcli main.go

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bfcli main.go

clean:
	rm bfcli
