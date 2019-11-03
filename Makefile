BINARY_NAME=goenvdir

build:
	go build -o $(BINARY_NAME) -v

test:
	go vet
	go test -v
