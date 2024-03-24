package main

import (
	"flag"
	"log"
	"net/http"
)

var (
	bindAddr  string
	routerURL string
	username  string
	password  string
)

func init() {
	flag.StringVar(&bindAddr, "bind-addr", "localhost:8888", "Address to bind to")
	flag.StringVar(&routerURL, "router-url", "http://localhost", "URL of the router")
	flag.StringVar(&username, "username", "admin", "Username for the router")
	flag.StringVar(&password, "password", "admin", "Password for the router")
}

var client *http.Client

func main() {
	flag.Parse()

	client = &http.Client{}

	http.HandleFunc("GET /records", Records)
	http.HandleFunc("POST /adjustendpoints", AdjustEndpoints)
	http.HandleFunc("POST /records", ApplyChanges)
	http.HandleFunc("GET /healthz", Healthz)

	log.Fatal(http.ListenAndServe(bindAddr, nil))
}
