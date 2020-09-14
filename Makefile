.PHONY: all build clean

all: build

build:
	cd cli && go build
	cd cmd/ezl && go build -ldflags "-s -w"
	cd cmd/ez && go build -ldflags "-s -w"
	cd cmd/ezt && go build -ldflags "-s -w"

clean:
	cd cmd/ezl && go clean
	cd cmd/ez && go clean
	cd cmd/ezt && go clean
