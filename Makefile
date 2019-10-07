.PHONY: test
default: test

.PHONY: build
build:
	go build

minitor-go:
	go build

.PHONY: run
run: minitor-go build
	./minitor-go -debug

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
