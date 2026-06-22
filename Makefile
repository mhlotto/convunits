.PHONY: build clean test

build:
	go build -o convunits ./cmd/convunits

test:
	go test ./...

clean:
	rm -f convunits
