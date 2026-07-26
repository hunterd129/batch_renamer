.PHONY: build clean run test install

BINARY_NAME=batch_renamer

build:
	go build -ldflags="-s -w" -o bin/$(BINARY_NAME) .

run: build./bin/$(BINARY_NAME)

test:
	go test -v ./...

clean:
	rm -f bin/*

install: build
	install -Dm755 bin/$(BINARY_NAME)
