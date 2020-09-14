.PHONY: all build clean

all: build

build:
	cd cli && go build
	cd cmd/ezl && go build -ldflags "-s -w"

clean:
	cd cmd/ezl && go clean
