.PHONY: lint
lint:
	@which golangci-lint >/dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.54.1
	golangci-lint run ./...

.PHONY: dev
dev:
	SADNESS_MYSQL_DATABASE=micro_post SADNESS_MYSQL_USER=ojisan SADNESS_MYSQL_PASSWORD=ojisan SADNESS_MYSQL_HOST=localhost:3306 go run main.go
