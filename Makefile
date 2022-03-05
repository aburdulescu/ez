BINDIR = bin/$(shell go env GOOS)/$(shell go env GOARCH)

.PHONY: build test clean update

build:
	mkdir -p $(BINDIR)
	cd cmn && go build
	cd ezt && go build
	cd swp && go build
	cd cadet && go build
	cd cadet/test && go build
	cd cmd/ez && go build
	cp cmd/ez/ez $(BINDIR)/
	cd cmd/ezl && go build
	cp cmd/ezl/ezl $(BINDIR)/
	cd cmd/ezt && go build
	cp cmd/ezt/ezt $(BINDIR)/
	cd cmd/ezs && go build
	cp cmd/ezs/ezs $(BINDIR)/

test:
	cd ezt && go test
	cd cmn && go test
	cd swp && go test
	cd cadet && go test
	cd cmd/ezl && go test
	cd cmd/ez && go test
	cd cmd/ezt && go test
	cd cmd/ezs && go test

clean:
	rm -rf $(BINDIR)
	cd cadet/test && go clean
	cd cmd/ezl && go clean
	cd cmd/ez && go clean
	cd cmd/ezt && go clean
	cd cmd/ezs && go clean
	find -type f -name "f*B" | xargs rm -f

update:
	cd cmn && go get -u && go mod tidy
	cd swp && go get -u && go mod tidy
	cd ezt && go get -u && go mod tidy
	cd cadet && go get -u && go mod tidy
	cd cadet/test && go get -u && go mod tidy
	cd cmd/ez && go get -u && go mod tidy
	cd cmd/ezl && go get -u && go mod tidy
	cd cmd/ezt && go get -u && go mod tidy
	cd cmd/ezs && go get -u && go mod tidy
