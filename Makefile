BINARY=socnotes

.PHONY: build run test clean

build:
	go build -ldflags="-s -w" -o $(BINARY) .

run:
	go run .

test:
	go test ./... -v

clean:
	rm -f $(BINARY)
