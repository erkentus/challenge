# Challenge

## Requires Go 1.7

## Run
`go run main.go`

## Assumptions:

1. Query parameters other than `u=...` are ignored
2. Query parameters are absolute valid URLs, others are treated as invalid urls and ignored
3. `GET` requets are used to fetch the list of numbers from the provided URLs
4. 200 status code is expected from the endpoint 
5. Slight overhead of data postprocessing(merging + deduplication) is fixed at 25ms - this of course very much depends on the testing environment and amount of passed data 
6. 500ms excludes connection latency 
7. test/testserver.go was slightly altered to return random waiting in range `0..750` instead of `0..550`. And listen on port `8079`.

## Instructions

1. Running unit tests: `go test`
2. Run e2e tests with either `make e2e` or via manually starting
2.1 `go run -race main.go&` - main server
2.2 `go run -race test/testserver.go&` - endpoints providing data
2.3.`go run -race e2e/test.go` - e2e tests/benchmarking
 