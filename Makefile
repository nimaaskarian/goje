all: coverage.out bin/tomc bin/tom
SRC = $(wildcard requests/*.go timer/*.go)
SRC_SERVER = $(wildcard cmd/server/*.go)
SRC_CLIENT = $(wildcard cmd/client/*.go)

coverage.out: $(shell find -name "*.go" -type f)
	go test ./... -coverprofile=coverage.out

bin/tomc: $(SRC_CLIENT) $(SRC)
	cd cmd/client && go build
	mv cmd/client/client bin/tomc

bin/tom: $(SRC_SERVER) $(SRC)
	cd cmd/server && go build
	mv cmd/server/server bin/tom
