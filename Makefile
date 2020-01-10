DOCKER_TAG ?= minitor-go-${USER}
GIT_TAG_NAME := $(shell git tag -l --contains HEAD)
GIT_SHA := $(shell git rev-parse HEAD)
VERSION := $(if $(GIT_TAG_NAME),$(GIT_TAG_NAME),$(GIT_SHA))

.PHONY: all
all: minitor-linux-amd64 minitor-linux-arm minitor-linux-arm64

.PHONY: default
default: test

.PHONY: build
build: minitor

minitor:
	@echo Version: $(VERSION)
	go build -ldflags '-X "main.version=${VERSION}"' -o minitor

.PHONY: run
run: minitor build
	./minitor -debug

.PHONY: run-metrics
run-metrics: minitor build
	./minitor -debug -metrics

.PHONY: test
test:
	go test -coverprofile=coverage.out
	@echo
	go tool cover -func=coverage.out
	@echo
	@# Check min coverage percentage
	@go tool cover -func=coverage.out | awk -v target=80.0% \
		'/^total:/ { print "Total coverage: " $$3 " Minimum coverage: " target; if ($$3+0.0 >= target+0.0) print "ok"; else { print "fail"; exit 1; } }'

# Installs pre-commit hooks
.PHONY: install-hooks
install-hooks:
	pre-commit install --install-hooks

# Checks files for encryption
.PHONY: check
check:
	pre-commit run --all-files

.PHONY: clean
clean:
	rm -f ./minitor
	rm -f ./minitor-linux-*
	rm -f ./minitor-darwin-amd64
	rm -f ./coverage.out

.PHONY: docker-build
docker-build:
	docker build -f ./Dockerfile.multi-stage -t $(DOCKER_TAG)-linux-amd64 .

.PHONY: docker-run
docker-run: docker-build
	docker run --rm -v $(shell pwd)/config.yml:/root/config.yml $(DOCKER_TAG)

## Multi-arch targets

# Arch specific go build targets
minitor-darwin-amd64:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 \
		   go build -ldflags '-X "main.version=${VERSION}"' -a -installsuffix nocgo \
		   -o minitor-darwin-amd64

minitor-linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		   go build -ldflags '-X "main.version=${VERSION}"' -a -installsuffix nocgo \
		   -o minitor-linux-amd64

minitor-linux-arm:
	GOOS=linux GOARCH=arm CGO_ENABLED=0 \
		   go build -ldflags '-X "main.version=${VERSION}"' -a -installsuffix nocgo \
		   -o minitor-linux-arm

minitor-linux-arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
		   go build -ldflags '-X "main.version=${VERSION}"' -a -installsuffix nocgo \
		   -o minitor-linux-arm64

# Arch specific docker build targets
.PHONY: docker-build-arm
docker-build-arm: minitor-linux-arm
	docker build --build-arg REPO=arm32v7 --build-arg ARCH=arm . -t ${DOCKER_TAG}-linux-arm

.PHONY: docker-build-arm
docker-build-arm64: minitor-linux-arm64
	docker build --build-arg REPO=arm64v8 --build-arg ARCH=arm64 . -t ${DOCKER_TAG}-linux-arm64

# Cross run on host architechture
.PHONY: docker-run-arm
docker-run-arm: docker-build-arm
	docker run --rm -v /var/run/docker.sock:/var/run/docker.sock --name $(DOCKER_TAG)-run ${DOCKER_TAG}-linux-arm

.PHONY: docker-run-arm64
docker-run-arm64: docker-build-arm64
	docker run --rm -v /var/run/docker.sock:/var/run/docker.sock --name $(DOCKER_TAG)-run ${DOCKER_TAG}-linux-arm64
