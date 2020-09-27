.PHONY: build test clean docker

build:
	cd cli && go build
	cd chunks && go build
	cd hash && go build
	cd ezs && make && go build
	cd ezt && go build
	cd cmd/ez && go build -ldflags "-s -w"
	cd cmd/ezl && go build -ldflags "-s -w"
	cd cmd/ezt && go build -ldflags "-s -w"
	cd cmd/ezs && go build -ldflags "-s -w"

test:
	cd cli && go test
	cd ezt && go test
	cd ezs && go test
	cd chunks && go test
	cd hash && go test
	cd cmd/ezl && go test
	cd cmd/ez && go test
	cd cmd/ezt && go test
	cd cmd/ezs && go test

clean:
	cd cmd/ezl && go clean
	cd cmd/ez && go clean
	cd cmd/ezt && go clean
	cd cmd/ezs && go clean
	find -type f -name "f*B" | xargs rm -f

update:
	cd cli && go mod tidy && go get -u
	cd chunks && go mod tidy && go get -u
	cd hash && go mod tidy && go get -u
	cd ezs && go mod tidy && go get -u
	cd ezt && go mod tidy && go get -u
	cd cmd/ez && go mod tidy && go get -u
	cd cmd/ezl && go mod tidy && go get -u
	cd cmd/ezt && go mod tidy && go get -u
	cd cmd/ezs && go mod tidy && go get -u

docker: clean
	docker build -t ez_base:latest .
