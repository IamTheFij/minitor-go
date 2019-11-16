.PHONY: all
DOCKER_TAG ?= minitor-go-${USER}

.PHONY: default
default: test

.PHONY: build
build:
	go build

minitor-go:
	go build

.PHONY: run
run: minitor-go build
	./minitor-go -debug

.PHONY: run-metrics
run-metrics: minitor-go build
	./minitor-go -debug -metrics

.PHONY: test
test:
	go test -coverprofile=coverage.out
	@echo
	go tool cover -func=coverage.out
	@echo
	@# Check min coverage percentage
	@go tool cover -func=coverage.out | awk -v target=80.0% \
		'/^total:/ { print "Total coverage: " $$3 " Minimum coverage: " target; if ($$3+0.0 >= target+0.0) print "ok"; else { print "fail"; exit 1; } }'

.PHONY: clean
clean:
	rm -f ./minitor-go
	rm -f ./coverage.out

.PHONY: docker-build
docker-build:
	docker build -f ./Dockerfile.multi-stage -t $(DOCKER_TAG) .

.PHONY: docker-run
docker-run: docker-build
	docker run --rm -v $(shell pwd)/config.yml:/root/config.yml $(DOCKER_TAG)
