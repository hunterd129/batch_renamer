.PHONY: build build-win build-all clean install

BINARY_NAME=batch_renamer

build:
	go build -ldflags="-s -w" -o bin/$(BINARY_NAME) .

build-win:
	GOOS=windows GOARCH=amd64 go build -ldflags"-s -w" -o bin/$(BINARY_NAME).exe .

build-all: build build-win

clean:
	rm -f bin/*

install: build
	install -Dm755 bin/$(BINARY_NAME)
