default: build

.PHONY: build
build:
	go build

.PHONY: run
run:
	go run main.go

.PHONY: test
test:
	go test -race -cover ./...