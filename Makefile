all: coverage.out tom
SRC = $(shell find -name "*.go" -type f)

coverage.out: $(SRC)
	go test ./... -coverprofile=coverage.out

tom: $(SRC)
	go build
