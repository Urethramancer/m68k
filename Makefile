# Makefile for m68k

.PHONY: all build clean test fmt vet

BINS := bin/asm68 bin/dis68 bin/run68

all: build

build: $(BINS)

bin/asm68: 
	go build -o $@ ./cmd/asm68

bin/dis68:
	go build -o $@ ./cmd/dis68

bin/run68:
	go build -o $@ ./cmd/run68

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	rm -f $(BINS)
