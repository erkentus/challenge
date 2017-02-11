package main

import "testing"
import "time"
import "net/http/httptest"
import "net/http"
import "math/rand"
import "context"
import "net/url"
import "encoding/json"

func TestFetchData(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	timeout := 500 * time.Millisecond

	aggr := &Aggregator{
		data: &Data{make([]int, 0)},
	}

	wrong := "127.0.0.1:2222"
	wrongURL, _ := url.Parse(wrong)
	if aggr.fetchData(context.TODO(), wrongURL) != errServerUnavailable {
		t.Errorf("Request should fail")
	}
	// mock servers
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(timeout + time.Duration(rand.Intn(50)))
		w.Write([]byte("Ignore me"))
	}))
	defer slow.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	slowURL, _ := url.Parse(slow.URL)
	if aggr.fetchData(ctx, slowURL) != errRequestTookTooLong {
		t.Errorf("Request should take too long")
	}

	unavailable := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("Something went wrong"))
	}))
	defer unavailable.Close()

	unavailableURL, _ := url.Parse(unavailable.URL)
	if aggr.fetchData(context.TODO(), unavailableURL) != errUnexpectedResponse {
		t.Errorf("Request should fail")
	}

	stupid := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("This is not JSON"))
	}))
	defer stupid.Close()

	stupidURL, _ := url.Parse(stupid.URL)
	if aggr.fetchData(context.TODO(), stupidURL) != errUnexpectedResponse {
		t.Errorf("Request should fail")
	}

	normal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&Data{Numbers: []int{1, 2, 3, 4, 5}})
	}))
	defer normal.Close()

	normalURL, _ := url.Parse(normal.URL)
	if aggr.fetchData(context.TODO(), normalURL) != nil {
		t.Errorf("Request should not fail")
	}
	if !compSlice(aggr.data.Numbers, []int{1, 2, 3, 4, 5}) {
		t.Errorf("Incorrect data inserted, got: %v, expected: %v", aggr.data.Numbers,
			[]int{1, 2, 3, 4, 5})
	}
}

func TestToSet(t *testing.T) {
	data := &Data{
		Numbers: []int{},
	}

	a := &Aggregator{
		data: data,
	}

	a.data.Numbers = []int{}

	if !compSlice(a.set().Numbers, []int{}) {
		t.Errorf("aggregator.set() error, got: %v, expected: %v", a.set().Numbers, []int{})
	}

	a.data.Numbers = []int{1, 2, 3, 4, 5}

	if !compSlice(a.set().Numbers, []int{1, 2, 3, 4, 5}) {
		t.Errorf("aggregator.set() error, got: %v, expected: %v", a.set().Numbers, []int{1, 2, 3, 4, 5})
	}

	a.data.Numbers = []int{1, 1, 3, 4, 5}

	if !compSlice(a.set().Numbers, []int{1, 3, 4, 5}) {
		t.Errorf("aggregator.set() error, got: %v, expected: %v", a.set().Numbers, []int{1, 3, 4, 5})
	}

	a.data.Numbers = []int{300, 4, 500, 500, 4}

	if !compSlice(a.set().Numbers, []int{4, 300, 500}) {
		t.Errorf("aggregator.set() error, got: %v, expected: %v", a.set().Numbers, []int{4, 300, 500})
	}
}

func compSlice(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
