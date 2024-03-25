tests/lint:
	docker run --rm -v $(PWD):/src -w /src golangci/golangci-lint:latest golangci-lint run --color always

tests/unit:
	docker run --rm -v $(PWD):/src -w /src golang:latest go test -cover ./...
