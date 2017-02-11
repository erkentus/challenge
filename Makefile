.PHONY: test build e2e run_server run_endpoints

build:
	@go build -o numberizer

e2e: test run_server run_endpoints
	sleep 2
	@go run -race e2e/test.go
	@pgrep numberizer | xargs -I {} kill -s TERM {}
	@pgrep testserve | xargs -I {} kill -s TERM {}
	@rm -rf numberizer

test:
	@go test

run_server: build
	@./numberizer -race > /dev/null 2>&1 &

run_endpoints:
	@go run -race test/testserver.go > /dev/null 2>&1 &