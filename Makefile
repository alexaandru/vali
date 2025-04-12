all: fmt vulncheck lint test

fmt:
	@find -name "*.go"|xargs go tool -modfile=tools/go.mod gofumpt -extra -w
	@find -name "*.go"|xargs go tool -modfile=tools/go.mod goimports -w

vulncheck:
	@go tool -modfile=tools/go.mod govulncheck ./...

lint:
	@go tool -modfile=tools/go.mod golangci-lint config verify
	@go tool -modfile=tools/go.mod golangci-lint run
	@go tool -modfile tools/go.mod modernize -test ./...

test:
	@go test -vet=all -cover -covermode=atomic -coverprofile=unit.cov .

clean:
	@rm -rf unit.cov unit.svg
