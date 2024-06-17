GO := $(shell which go)
VERSION = 1.3

.PHONY: clean

all: protrans

protrans: cmd/protrans/main.go
	CGO_ENABLED=0 $(GO) build -ldflags "-X 'main.Version=${VERSION}'" -o $@ -v $^

test:
	CGO_ENABLED=0 $(GO) test -v ./...

clean:
	rm -fr protrans
