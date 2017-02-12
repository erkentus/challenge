# Challenge

## Requirements
- Go 1.7
- Cmake for Makefile - not required

## Run
`go run main.go`

This will start the service at `http://localhost:8080/numbers` 

## Assumptions:

1. Query parameters other than `u=...` are ignored
2. Query parameters are absolute valid URLs, others are treated as invalid urls and ignored
3. `GET` requets are used to fetch the list of numbers from the provided URLs
4. 200 status code is expected from the endpoint 
5. Slight overhead of data postprocessing(merging + deduplication) is fixed at 25ms - this of course very much depends on the testing environment and amount of passed data 
6. 500ms excludes initial client-server connection latency 
7. test/testserver.go was slightly altered to return random waiting in range `0..750` instead of `0..550`. And listen on port `8079`.
8. Reasonable number of endpoints (test for <= 1000) and payload sizes are assumed. 

## Instructions

Unit and end-to-end tests are provided. End to end test runs 10 requests (with 0..1000 random number of endpoints) one after another 
with estimated request time. Since tested on the same machine connection latency is neglected. 

Endpoints are the provided by slightly altered `testserver.go`

1. Running unit tests: `go test`
2. Run e2e tests with either `make e2e` or via manually starting

-  `go run main.go&` - main server

-  `go run test/testserver.go&` - endpoints providing data

-  `go run e2e/test.go` - e2e tests/benchmarking
 
Makefile is tested on OSX 10.12.3 but should work on modern Linux as well