# Makefile for m68k
FLAGS=-trimpath -ldflags "-w -s"
DEBUG_FLAGS=-gcflags "all=-N -l"

.PHONY: all build clean test fmt vet

BINS := bin/asm68 bin/dis68 bin/run68
BINS-debug := bin/asm68-debug bin/dis68-debug bin/run68-debug


all: build

build: $(BINS)
debug: $(BINS-debug)

bin/asm68:
	go build ${FLAGS} -o $@ ./cmd/asm68

bin/asm68-debug:
	go build ${DEBUG_FLAGS} -o $@ ./cmd/asm68

bin/dis68:
	go build ${FLAGS} -o $@ ./cmd/dis68

bin/dis68-debug:
	go build ${DEBUG_FLAGS} -o $@ ./cmd/dis68

bin/run68:
	go build ${FLAGS} -o $@ ./cmd/run68

bin/run68-debug:
	go build ${DEBUG_FLAGS} -o $@ ./cmd/run68

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	rm -f $(BINS) $(BINS-debug)
