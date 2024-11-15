DOCKER_TAG ?= minitor-go-${USER}
VERSION ?= $(shell git describe --tags --dirty)
GOFILES = *.go go.mod go.sum
# Multi-arch targets are generated from this
TARGET_ALIAS = minitor-linux-amd64 minitor-linux-arm minitor-linux-arm64 minitor-darwin-amd64
TARGETS = $(addprefix dist/,$(TARGET_ALIAS))
#
# Default make target will run tests
.DEFAULT_GOAL = test

# Build all static Minitor binaries
.PHONY: all
all: $(TARGETS)

# Build all static Linux Minitor binaries. Used in Docker images
.PHONY: all-linux
all-linux: $(filter dist/minitor-linux-%,$(TARGETS))

# Build minitor for the current machine
minitor: $(GOFILES)
	@echo Version: $(VERSION)
	go build -ldflags '-X "main.version=${VERSION}"' -o minitor

.PHONY: build
build: minitor

# Run minitor for the current machine
.PHONY: run
run: minitor
	./minitor -debug

.PHONY: run-metrics
run-metrics: minitor
	./minitor -debug -metrics

# Run all tests
.PHONY: test
test:
	go test -coverprofile=coverage.out
	go tool cover -func=coverage.out
	@go tool cover -func=coverage.out | awk -v target=80.0% \
		'/^total:/ { print "Total coverage: " $$3 " Minimum coverage: " target; if ($$3+0.0 >= target+0.0) print "ok"; else { print "fail"; exit 1; } }'

# Installs pre-commit hooks
.PHONY: install-hooks
install-hooks:
	pre-commit install --install-hooks

# Runs pre-commit checks on files
.PHONY: check
check:
	pre-commit run --all-files

.PHONY: clean
clean:
	rm -f ./minitor
	rm -f ./coverage.out
	rm -fr ./dist

.PHONY: docker-build
docker-build:
	docker build -f ./Dockerfile.multi-stage -t $(DOCKER_TAG)-linux-amd64 .

.PHONY: docker-run
docker-run: docker-build
	docker run --rm -v $(shell pwd)/sample-config.hcl:/root/config.hcl $(DOCKER_TAG)

## Multi-arch targets
$(TARGETS): $(GOFILES)
	mkdir -p ./dist
	GOOS=$(word 2, $(subst -, ,$(@))) GOARCH=$(word 3, $(subst -, ,$(@))) CGO_ENABLED=0 \
		 go build -ldflags '-X "main.version=${VERSION}"' -a -installsuffix nocgo \
		 -o $@

.PHONY: $(TARGET_ALIAS)
$(TARGET_ALIAS):
	$(MAKE) $(addprefix dist/,$@)

# Arch specific docker build targets
.PHONY: docker-build-arm
docker-build-arm: dist/minitor-linux-arm
	docker build --platform linux/arm . -t ${DOCKER_TAG}-linux-arm

.PHONY: docker-build-arm64
docker-build-arm64: dist/minitor-linux-arm64
	docker build  --platform linux/arm64 . -t ${DOCKER_TAG}-linux-arm64

# Cross run on host architechture
.PHONY: docker-run-arm
docker-run-arm: docker-build-arm
	docker run --rm -v /var/run/docker.sock:/var/run/docker.sock --name $(DOCKER_TAG)-run ${DOCKER_TAG}-linux-arm

.PHONY: docker-run-arm64
docker-run-arm64: docker-build-arm64
	docker run --rm -v /var/run/docker.sock:/var/run/docker.sock --name $(DOCKER_TAG)-run ${DOCKER_TAG}-linux-arm64
