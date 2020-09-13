.PHONY: all build clean

all: build

build:
	cd cmd/ezl && go build

clean:
	cd cmd/ezl && go clean
