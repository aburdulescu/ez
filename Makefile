BINDIR = bin_$(shell go env GOOS)_$(shell go env GOARCH)

.PHONY: build test clean update

build:
	mkdir -p $(BINDIR)
	cd cmn && go build
	cd ezt && go build
	cd swp && go build
	cd cadet && go build
	cd cadet/test && go build -ldflags "-s -w"
	cd cmd/ez && go build -ldflags "-s -w"
	cp cmd/ez/ez $(BINDIR)/
	cd cmd/ezl && go build -ldflags "-s -w"
	cp cmd/ezl/ezl $(BINDIR)/
	cd cmd/ezt && go build -ldflags "-s -w"
	cp cmd/ezt/ezt $(BINDIR)/
	cd cmd/ezs && go build -ldflags "-s -w"
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
	cd cmn && go mod tidy && go get -u
	cd swp && go mod tidy && go get -u
	cd ezt && go mod tidy && go get -u
	cd cadet && go mod tidy && go get -u
	cd cadet/test && go mod tidy && go get -u
	cd cmd/ez && go mod tidy && go get -u
	cd cmd/ezl && go mod tidy && go get -u
	cd cmd/ezt && go mod tidy && go get -u
	cd cmd/ezs && go mod tidy && go get -u
