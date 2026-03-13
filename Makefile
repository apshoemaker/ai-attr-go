.PHONY: build test vet clean install

BINARY := ai-attr
VERSION := 0.2.0
TEST := foo
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/ai-attr

test:
	go test ./...

test-v:
	go test -v ./...

vet:
	go vet ./...

clean:
	rm -f $(BINARY)

install: build
	cp $(BINARY) $(GOPATH)/bin/ 2>/dev/null || cp $(BINARY) $(HOME)/go/bin/
