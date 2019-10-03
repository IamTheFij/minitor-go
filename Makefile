.PHONY: test
default: test

.PHONY: build
build:

minitor-go:
	go build

.PHONY: run
run: minitor-go
	./minitor-go

.PHONY: test
test:
	go test -coverprofile=coverage.out
	go tool cover -func=coverage.out

.PHONY: clean
clean:
	rm -f ./minitor-go
	rm -f ./coverage.out
