GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
MAIN_FILE=./main.go
BINARY_NAME=nesgo

test:
	$(GOTEST) -v ./...

coverage:
	$(GOTEST) -v ./... -cover -coverprofile=coverage.out

coverage-html:
	make coverage
	go tool cover -html=coverage.out

build:
	 $(GOBUILD) -o $(BINARY_NAME) -v

run:
	$(GORUN) $(MAIN_FILE)

build-static-linux:
	CGO_ENABLED=1 CC=gcc GOOS=linux GOARCH=amd64 $(GOBUILD) -tags static -ldflags "-s -w"