package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const (
	server         = `http://localhost:8080/numbers`
	endpointPrimes = `http://localhost:8079/primes`
	endpointRand   = `http://localhost:8079/rand`
	endpointFibo   = `http://localhost:8079/fibo`
	endpointOdd    = `http://localhost:8079/odd`

	timeout = 0.5
)

type data struct {
	Numbers []int `json:"numbers"`
}

func main() {
	rand.Seed(time.Now().UnixNano())
	endpoints := []string{endpointPrimes, endpointFibo, endpointRand, endpointOdd}

	numQueries := 10

	for i := 0; i < numQueries; i++ {
		queryURL := fmt.Sprintf("%s?", server)
		numEndpoints := rand.Intn(1000)
		for j := 0; j < numEndpoints; j++ {
			endpoint := rand.Intn(4)
			queryURL = fmt.Sprintf("%su=%s&", queryURL, endpoints[endpoint])
		}

		//request start time
		//here latency and request construction
		//are negligible
		start := time.Now()

		res, err := http.Get(queryURL)
		if err != nil {
			log.Fatalf("e2e error: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			log.Fatalf("e2e error. Status code: %d", res.StatusCode)
		}

		var expectedData data
		err = json.NewDecoder(res.Body).Decode(&expectedData)
		if err != nil {
			log.Fatalf("e2e error: %v", err)
		}

		elapsed := time.Since(start)
		if elapsed.Seconds() >= timeout {
			log.Fatalf("e2e test 500ms exceeded!")
		}
		log.Printf("E2E test successfully passed!")
		log.Printf("Number of endpoints: %d", numEndpoints)
		log.Printf("Elapsed time: %v", elapsed)
		log.Printf("___________")
	}
}
