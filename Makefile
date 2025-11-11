all: fmt vulncheck lint test

fmt:
	@go fmt ./...
	@go tool -modfile=tools/go.mod goimports -l -w .
	@go run mvdan.cc/gofumpt@v0.8.0 -l -w -extra .

vulncheck:
	@go tool -modfile=tools/go.mod govulncheck ./...

lint:
	@go tool -modfile=tools/go.mod golangci-lint config verify
	@go tool -modfile=tools/go.mod golangci-lint run

test:
	@go test -vet=all -cover -covermode=atomic -coverprofile=unit.cov .
	@go tool -modfile=tools/go.mod stampli -quiet -coverage=$$(go tool cover -func=unit.cov|tail -n1|tr -s "\t"|cut -f3|tr -d "%")

clean:
	@rm -rf unit.cov unit.svg
