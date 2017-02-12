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
const overhead = 25 //responsible for the last step of merging the results
const queryKey = "u"
const timeout = (500 - overhead) * time.Millisecond //subtract the overhead to make sure 500ms requirement is met

var ( //custom errors
	errRequestTookTooLong = errors.New("Request took too long to finish")
	errUnexpectedResponse = errors.New("Unexpected response")
	errUnknown            = errors.New("Unknown error")
	errServerUnavailable  = errors.New("Endpoint is unavailable")
)

// Aggregator is a core component control mechanism - thread safe
// protects from data races
type Aggregator struct {
	data *Data
	sync.Mutex
	sync.WaitGroup
}

// merge thread safe method to merge new set of numbers into data field
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
	// context here is used in conjuction with requests
	// to kill timed out requests and limit the lifespan of this function
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
				err := aggr.fetchData(ctx, endpoint) //fetch the data from endpoint
				if err != nil {
					log.Printf("[Error] Fetching data from %s, error: %v", endpoint.String(), err)
				}
			}()
		}
		aggr.Wait()
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done(): //timeout
	case <-done: //all requests were either finished or failed
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(200)
	json.NewEncoder(res).Encode(aggr.set())
}

// fetchData method is responsible for getting the data and adding it to
// the data field via method receiver. It also returns custom error for ease of testing
func (a *Aggregator) fetchData(ctx context.Context, endpoint *url.URL) error {
	req, err := http.NewRequest("GET", endpoint.String(), nil)
	if err != nil {
		return errUnknown
	}
	// pass the context to the http client
	res, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return errRequestTookTooLong
		}
		return errServerUnavailable
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return errUnexpectedResponse
	}

	var numbers Data
	if err = json.NewDecoder(res.Body).Decode(&numbers); err != nil {
		return errUnexpectedResponse
	}

	log.Printf("Received: %v from %s", numbers, endpoint.String())
	a.merge(numbers)

	return nil
}
