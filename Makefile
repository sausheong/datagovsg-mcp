BINARY := datagovsg-mcp
PORT   := 8080

.PHONY: build test run run-http clean

build:
	go build -o $(BINARY) .

test:
	go test ./...

run: build
	./$(BINARY)

run-http: build
	./$(BINARY) --http :$(PORT)

clean:
	rm -f $(BINARY)
