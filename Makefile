GO := $(shell which go)
VERSION = 1.3

.PHONY: clean

all: protrans

protrans: cmd/protrans/main.go
	$(GO) build -ldflags "-X 'main.Version=${VERSION}'" -o $@ -v $^

test:
	$(GO) test -v ./...

clean:
	rm -fr protrans
