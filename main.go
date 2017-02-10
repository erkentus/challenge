package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

const listenPort = ":8080"

// Result the final result to be returned
type Result struct {
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
	result := Result{
		Numbers: []int{1, 2, 3, 4, 5},
	}
	jsonBlob, err := json.Marshal(result)
	if err != nil {
		res.WriteHeader(500)
		fmt.Fprintf(res, "Failed to json encode results: %v", err)
		return
	}
	res.WriteHeader(200)
	res.Write(jsonBlob)
}

func fetchURL(link url.URL) {

}
