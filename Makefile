.PHONY: lint
lint:
	golangci-lint run ./...


.PHONY: test
test:
	go clean -testcache
	go test ./... -race -covermode=atomic -coverprofile=coverage.out
