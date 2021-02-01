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
	curl https://sum.golang.org/lookup/github.com/buypal/oapi-go@v$(VERSION)

test:
	go test ./...

install:
	cd cmd/oapi && go install 

build:
	cd cmd/oapi && go build 