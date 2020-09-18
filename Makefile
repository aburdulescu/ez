.PHONY: build test clean

build:
	cd cli && go build
	cd ezt && go build
	cd ezs && make && go build
	cd cmd/ezl && go build -ldflags "-s -w"
	cd cmd/ez && go build -ldflags "-s -w"
	cd cmd/ezt && go build -ldflags "-s -w"
	cd cmd/ezs && go build -ldflags "-s -w"

test:
	cd cli && go test
	cd ezt && go test
	cd ezs && go test
	cd cmd/ezl && go test
	cd cmd/ez && go test
	cd cmd/ezt && go test
	cd cmd/ezs && go test

clean:
	cd cmd/ezl && go clean
	cd cmd/ez && go clean
	cd cmd/ezt && go clean
	cd cmd/ezs && go clean
