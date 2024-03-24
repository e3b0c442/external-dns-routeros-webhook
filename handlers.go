package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
)

func Records(w http.ResponseWriter, r *http.Request) {
	reqURL, err := url.JoinPath(routerURL, "/rest/ip/dns/static")
	if err != nil {
		slog.Error("Failed to join URL", "error", err)
		http.Error(w, "Failed to join URL", http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		slog.Error("Failed to create request", "error", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Failed to get records", "error", err)
		http.Error(w, "Failed to get records", http.StatusInternalServerError)
		return
	}
	if resp.StatusCode != http.StatusOK {
		buf := &bytes.Buffer{}
		if _, err := buf.ReadFrom(resp.Body); err != nil {
			slog.Error("Failed to read response body", "error", err)
			http.Error(w, "Failed to read response body on error", http.StatusInternalServerError)
			return
		}
		slog.Error("Failed to get records", "status", resp.Status, "body", buf.String())
		http.Error(w, "Failed to get records", http.StatusInternalServerError)
		return
	}

	var rawRecords []Record
	if err := json.NewDecoder(resp.Body).Decode(&rawRecords); err != nil {
		slog.Error("Failed to decode records", "error", err)
		http.Error(w, "Failed to decode records", http.StatusInternalServerError)
		return
	}

	endpoints, err := recordsToEndpoints(rawRecords)
	if err != nil {
		slog.Error("Failed to convert records to endpoints", "error", err)
		http.Error(w, "Failed to convert records to endpoints", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(endpoints); err != nil {
		w.Header().Del("Content-Type")
		slog.Error("Failed to encode endpoints", "error", err)
		http.Error(w, "Failed to encode endpoints", http.StatusInternalServerError)
	}
}

func AdjustEndpoints(w http.ResponseWriter, r *http.Request) {
	// ...
}

func ApplyChanges(w http.ResponseWriter, r *http.Request) {
	// ...
}

func Healthz(w http.ResponseWriter, r *http.Request) {
	// ...
}
