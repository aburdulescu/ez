.PHONY: build test clean docker

build:
	cd cmn && go build
	cd ezt && go build
	cd swp && go build
	cd cmd/ez && go build -ldflags "-s -w"
	cd cmd/ezl && go build -ldflags "-s -w"
	cd cmd/ezt && go build -ldflags "-s -w"
	cd cmd/ezs && go build -ldflags "-s -w"

test:
	cd ezt && go test
	cd cmn && go test
	cd swp && go test
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
	cd cmn && go mod tidy && go get -u
	cd swp && go mod tidy && go get -u
	cd ezt && go mod tidy && go get -u
	cd cmd/ez && go mod tidy && go get -u
	cd cmd/ezl && go mod tidy && go get -u
	cd cmd/ezt && go mod tidy && go get -u
	cd cmd/ezs && go mod tidy && go get -u

docker: clean
	docker build -t ez_base:latest .
