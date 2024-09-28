package main

import (
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
)

var (
	debug           bool
	bindAddr        string
	routerURL       string
	username        string
	password        string
	metricsBindAddr string
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
	flag.StringVar(&metricsBindAddr, "metrics-addr", defaultEnv("WEBHOOK_METRICS_BIND_ADDR", "localhost:8080"), "Address for metrics and health checks")
}

var client *RouterOSAPIClient

func main() {
	flag.Parse()

	if debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	client = NewRouterOSAPIClient(username, password, routerURL)

	http.HandleFunc("GET /", GetDomainFilter)
	http.HandleFunc("GET /records", Records)
	http.HandleFunc("POST /adjustendpoints", AdjustEndpoints)
	http.HandleFunc("POST /records", ApplyChanges)
	http.HandleFunc("GET /healthz", Healthz)

	metricsServeMux := http.NewServeMux()
	metricsServeMux.HandleFunc("GET /healthz", Healthz)
	metricsListener := &http.Server{
		Addr:    metricsBindAddr,
		Handler: metricsServeMux,
	}
	slog.Info("Metrics listening on " + metricsBindAddr)
	go metricsListener.ListenAndServe()

	slog.Info("Listening on " + bindAddr)
	log.Fatal(http.ListenAndServe(bindAddr, nil))

}
