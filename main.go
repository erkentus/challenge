package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const listenPort = ":8080"
const queryKey = "u"

// Aggregator the final result to be returned
type Aggregator struct {
	Numbers []int `json:"numbers"`
	mtx     sync.Mutex
	wg      sync.WaitGroup
}

// Response received from the endpoint
type Response struct {
	res *http.Response
	err error
}

func main() {
	listen()
}

func listen() {
	http.HandleFunc("/numbers", handler)
	log.Fatal(http.ListenAndServe(listenPort, nil))
}

func handler(res http.ResponseWriter, req *http.Request) {
	aggr := &Aggregator{
		Numbers: make([]int, 0),
		wg:      sync.WaitGroup{},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	values := req.URL.Query()[queryKey]
	for _, val := range values {
		endpoint, err := url.ParseRequestURI(val)
		if err != nil {
			log.Printf("[Error] Passed URL cannot be parsed: %v\n", err)
			continue
		}
		aggr.wg.Add(1)
		go fetchURL(ctx, aggr, endpoint)
	}
	aggr.wg.Wait()

	jsonBlob, err := json.Marshal(aggr)
	if err != nil {
		res.WriteHeader(500)
		fmt.Fprintf(res, "Failed to json encode results: %v", err)
		return
	}
	res.WriteHeader(200)
	res.Write(jsonBlob)
}

func fetchURL(ctx context.Context, aggr *Aggregator, endpoint *url.URL) {
	defer aggr.wg.Done()
	tr := &http.Transport{}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", endpoint.String(), nil)
	if err != nil {
		log.Printf("[Error] Endpoint request creation failed: %v\n", err)
		return
	}

	reqTracker := make(chan Response, 1)

	go func() {
		res, err := client.Do(req)
		reqTracker <- Response{res, err}
	}()

	select {
	case <-ctx.Done():
		tr.CancelRequest(req)
		<-reqTracker
		log.Printf("Cancelling request due to timeout")
		return
	case resp := <-reqTracker:
		if resp.err != nil {
			log.Printf("[Error] Failed to make a request %v\n", resp.err)
			return
		}
		defer resp.res.Body.Close()
		body, err := ioutil.ReadAll(resp.res.Body)
		if err != nil {
			log.Printf("[Error] Unexpected response from endpoint: %s, error: %v\n", endpoint.String(), err)
			return
		}
		var numbers Aggregator
		if err = json.Unmarshal(body, numbers); err != nil {
			log.Printf("[Error] Failed to parse JSON response: %s\n", string(body))
			return
		}
		log.Println("Received: ", numbers)
	}
}
