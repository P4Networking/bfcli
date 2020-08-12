.PHONY: all clean

all: build

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pisc-cli main.go

clean:
	rm pisc-cli
