package main

import (
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
)

var (
	debug     bool
	bindAddr  string
	routerURL string
	username  string
	password  string
)

func defaultEnv(key, def string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return def
}

func init() {
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.StringVar(&bindAddr, "bind-addr", defaultEnv("WEBHOOK_BIND_ADDR", "localhost:8888"), "Address to bind to")
	flag.StringVar(&routerURL, "router-url", defaultEnv("WEBHOOK_ROUTER_URL", "http://192.168.88.1"), "URL of the router")
	flag.StringVar(&username, "router-username", defaultEnv("WEBHOOK_ROUTER_USERNAME", "admin"), "Username for the router")
	flag.StringVar(&password, "router-password", defaultEnv("WEBHOOK_ROUTER_PASSWORD", ""), "Password for the router")
}

var client *http.Client

func main() {
	flag.Parse()

	if debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	client = &http.Client{}

	http.HandleFunc("GET /", GetDomainFilter)
	http.HandleFunc("GET /records", Records)
	http.HandleFunc("POST /adjustendpoints", AdjustEndpoints)
	http.HandleFunc("POST /records", ApplyChanges)
	http.HandleFunc("GET /healthz", Healthz)

	log.Fatal(http.ListenAndServe(bindAddr, nil))
}
