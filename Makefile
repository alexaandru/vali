all: lint test

test:
	@go test -vet=all -cover -covermode=atomic -coverprofile=unit.cov .

check_lint:
	@golangci-lint version > /dev/null 2>&1 || \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint: check_lint
	@golangci-lint run

clean:
	@rm -rf unit.cov unit.svg
