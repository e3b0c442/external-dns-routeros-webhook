package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"

	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
)

var negotiatedMediaType string

// GetDomainFilter returns the domain filter. As the RouterOS DNS cache does
// not have zones, we return the empty filter and rely on the filter set in
// external-dns.
func GetDomainFilter(w http.ResponseWriter, r *http.Request) {
	slog.Debug("GET /", "request_headers", r.Header)
	negotiatedMediaType = r.Header.Get("Accept")

	w.Header().Set("Content-Type", negotiatedMediaType)

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(endpoint.NewDomainFilter(nil)); err != nil {
		slog.Error("Failed to encode domain filter", "error", err)
		http.Error(w, "Failed to encode domain filter", http.StatusInternalServerError)
		return
	}
	slog.Debug("GET /", "response_headers", w.Header(), "response_body", buf.String())

	w.Write(buf.Bytes())
}

func Records(w http.ResponseWriter, r *http.Request) {
	slog.Debug("GET /records", "request_headers", r.Header)

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
	slog.Debug("GET /records", "routeros_request_url", req.URL.String(), "routeros_request_headers", req.Header)

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Failed to get records", "error", err)
		http.Error(w, "Failed to get records", http.StatusInternalServerError)
		return
	}
	buf := &bytes.Buffer{}
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		slog.Error("Failed to read response body", "error", err)
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}

	slog.Debug("GET /records", "routeros_response_headers", resp.Header, "routeros_response_status", resp.Status, "routeros_response_body", buf.String())

	if resp.StatusCode != http.StatusOK {
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

	slog.Debug("GET /records", "routeros_records", rawRecords)

	endpoints, err := recordsToEndpoints(rawRecords)
	if err != nil {
		slog.Error("Failed to convert records to endpoints", "error", err)
		http.Error(w, "Failed to convert records to endpoints", http.StatusInternalServerError)
		return
	}

	buf.Reset()
	if err := json.NewEncoder(buf).Encode(endpoints); err != nil {
		slog.Error("Failed to encode endpoints", "error", err)
		http.Error(w, "Failed to encode endpoints", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", negotiatedMediaType)
	slog.Debug("GET /records", "response_headers", w.Header(), "response_body", buf.String())

	w.Write(buf.Bytes())
}

func AdjustEndpoints(w http.ResponseWriter, r *http.Request) {
	buf := &bytes.Buffer{}
	if _, err := buf.ReadFrom(r.Body); err != nil {
		slog.Error("Failed to read request body", "error", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	slog.Debug("POST /adjustendpoints", "request_headers", r.Header, "request_body", buf.String())

	var endpoints []*endpoint.Endpoint

	if err := json.NewDecoder(buf).Decode(&endpoints); err != nil {
		slog.Error("Failed to decode endpoints", "error", err)
		http.Error(w, "Failed to decode endpoints", http.StatusInternalServerError)
		return
	}

	// no-op

	buf.Reset()
	if err := json.NewEncoder(buf).Encode(endpoints); err != nil {
		slog.Error("Failed to encode endpoints", "error", err)
		http.Error(w, "Failed to encode endpoints", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", negotiatedMediaType)
	slog.Debug("POST /adjustendpoints", "response_headers", w.Header(), "response_body", buf.String())

	w.Write(buf.Bytes())
}

func ApplyChanges(w http.ResponseWriter, r *http.Request) {
	buf := &bytes.Buffer{}
	if _, err := buf.ReadFrom(r.Body); err != nil {
		slog.Error("Failed to read request body", "error", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	slog.Debug("POST /records", "request_headers", r.Header, "request_body", buf.String())

	var changes *plan.Changes

	if err := json.NewDecoder(buf).Decode(&changes); err != nil {
		slog.Error("Failed to decode changes", "error", err)
		http.Error(w, "Failed to decode changes", http.StatusInternalServerError)
		return
	}

	// TODO implementation

	w.WriteHeader(http.StatusNoContent)
}
