.PHONY: default report clean-cache lint
.DEFAULT_GOAL := default

default: lint

report:
	@go test -json . > report.out
	@go test -coverprofile report.out .
	@go tool cover -html=report.out -o=report.html
	@golangci-lint run --issues-exit-code 0 --out-format checkstyle > checkstyle.xml

clean-cache:
	@go clean -testcache

lint:
	@go test -v
	@golangci-lint run
