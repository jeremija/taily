.PHONY: all
all: bin build

bin:
	mkdir -p bin/

.PHONY: build
build:
	go build -o bin/taily ./cmd

.PHONY: test
test:
	go test -race ./...

.PHONY: clean
clean:
	rm -rf bin/

