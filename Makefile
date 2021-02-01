GOCMD=go
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install
GOGENERATE=$(GOCMD) generate

.PHONY: all

all: test build

# make update-pkg-cache VERSION=1.12.4
update-pkg-cache:
	GOPROXY=https://proxy.golang.org GO111MODULE=on go get github.com/buypal/oapi-go@v$(VERSION)

test:
	go test ./...

install:
	cd cmd/oapi && go install 

build:
	cd cmd/oapi && go build 