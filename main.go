package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"
)

const listenPort = ":8080"
const overhead = 25
const queryKey = "u"
const timeout = (500 - overhead) * time.Millisecond

var ( //custom errors
	errRequestTookTooLong = errors.New("Request took too long to finish")
	errUnexpectedResponse = errors.New("Unexpected response")
	errUnknown            = errors.New("Unknown error")
	errServerUnavailable  = errors.New("Endpoint is unavailable")
)

// Aggregator control mechanism - thread safe
type Aggregator struct {
	data *Data
	sync.Mutex
	sync.WaitGroup
}

// merge thread safe method to merge new data into data field
func (a *Aggregator) merge(d Data) error {
	a.Lock()
	defer a.Unlock()

	a.data.Numbers = append(a.data.Numbers, d.Numbers...)

	return nil
}

// set deduplicates and sorts the data in ascending order
// to be called after all data is merged
func (a *Aggregator) set() *Data {
	a.Lock()
	defer a.Unlock()

	if len(a.data.Numbers) == 0 {
		return &Data{Numbers: []int{}}
	}
	sort.Ints(a.data.Numbers)

	set := []int{a.data.Numbers[0]}
	for _, num := range a.data.Numbers {
		if set[len(set)-1] != num {
			set = append(set, num)
		}
	}
	return &Data{
		Numbers: set,
	}
}

// Data json [de]serialization structure type
type Data struct {
	Numbers []int `json:"numbers"`
}

func main() {
	listen()
}

func listen() {
	http.HandleFunc("/numbers", handler)
	log.Fatal(http.ListenAndServe(listenPort, nil))
}

func handler(res http.ResponseWriter, req *http.Request) {
	log.Printf("Processing new request: %s", req.URL.String())

	aggr := &Aggregator{
		data: &Data{make([]int, 0)},
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		values := req.URL.Query()[queryKey]
		for _, val := range values {
			endpoint, err := url.ParseRequestURI(val)
			if err != nil {
				log.Printf("[Warning] Passed URL cannot be parsed: %v. Ignoring\n", err)
				continue
			}
			aggr.Add(1)
			go func() {
				defer aggr.Done()
				aggr.fetchData(ctx, endpoint)
			}()
		}
		aggr.Wait()
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		log.Printf("one of the requests timed out!")
	case <-done:
		log.Printf("all endpoints are processed!")
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(200)
	json.NewEncoder(res).Encode(aggr.set())
}

func (a *Aggregator) fetchData(ctx context.Context, endpoint *url.URL) error {
	req, err := http.NewRequest("GET", endpoint.String(), nil)
	if err != nil {
		log.Printf("[Error] Endpoint request creation failed: %v\n", err)
		return errUnknown
	}

	res, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("[Error] Request took too long: %s", endpoint.String())
			return errRequestTookTooLong
		}
		log.Printf("[Error] Failed to make a request %v\n", err)
		return errServerUnavailable
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Printf("[Error] Unexpected status from %s: %s", endpoint.String(), res.Status)
		return errUnexpectedResponse
	}

	var numbers Data
	if err = json.NewDecoder(res.Body).Decode(&numbers); err != nil {
		log.Printf("[Error] Failed to parse JSON response. Error: %v\n", err)
		return errUnexpectedResponse
	}

	log.Printf("Received: %v from %s", numbers, endpoint.String())
	a.merge(numbers)

	return nil
}
