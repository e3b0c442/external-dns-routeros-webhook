package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"

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

	records, err := client.ReadDNSStaticRecords()
	if err != nil {
		slog.Error("Failed to list records", "error", err)
		http.Error(w, "Failed to list records", http.StatusInternalServerError)
		return
	}

	endpoints, err := recordsToEndpoints(records)
	if err != nil {
		slog.Error("Failed to convert records to endpoints", "error", err)
		http.Error(w, "Failed to convert records to endpoints", http.StatusInternalServerError)
		return
	}

	buf := &bytes.Buffer{}
	json.NewEncoder(buf).Encode(endpoints)

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

	records, err := client.ReadDNSStaticRecords()
	if err != nil {
		slog.Error("Failed to list records", "error", err)
		http.Error(w, "Failed to list records", http.StatusInternalServerError)
		return
	}

	recordMap := make(map[string]map[string][]*Record)
	for _, record := range records {
		if _, ok := recordMap[record.Name]; !ok {
			recordMap[record.Name] = make(map[string][]*Record)
		}
		if _, ok := recordMap[record.Name][record.Type]; !ok {
			recordMap[record.Name][record.Type] = []*Record{record}
		} else {
			recordMap[record.Name][record.Type] = append(recordMap[record.Name][record.Type], record)
		}
	}

	for _, ep := range endpoints {
		if _, ok := recordMap[ep.DNSName]; !ok {
			continue
		}
		if _, ok := recordMap[ep.DNSName][ep.RecordType]; !ok {
			continue
		}
		for _, record := range recordMap[ep.DNSName][ep.RecordType] {
			ep.ProviderSpecific = endpoint.ProviderSpecific{
				{
					Name:  ".id",
					Value: record.ID,
				},
			}
		}
	}

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

	for _, create := range changes.Create {
		records, err := endpointToRecords(create)
		if err != nil {
			slog.Error("Failed to convert endpoint to records", "error", err)
			http.Error(w, "Failed to convert endpoint to records", http.StatusInternalServerError)
			return
		}

		for _, record := range records {
			if err := client.CreateDNSStaticRecord(record); err != nil {
				slog.Error("Failed to create record", "error", err)
				http.Error(w, "Failed to create record", http.StatusInternalServerError)
				return
			}
		}
	}

	for _, delete := range changes.Delete {
		records, err := endpointToRecords(delete)
		if err != nil {
			slog.Error("Failed to convert endpoint to records", "error", err)
			http.Error(w, "Failed to convert endpoint to records", http.StatusInternalServerError)
			return
		}

		for _, record := range records {
			if err := client.DeleteDNSStaticRecord(record.ID); err != nil {
				slog.Error("Failed to delete record", "error", err)
				http.Error(w, "Failed to delete record", http.StatusInternalServerError)
				return
			}
		}
	}

	for _, update := range changes.UpdateNew {
		records, err := endpointToRecords(update)
		if err != nil {
			slog.Error("Failed to convert endpoint to records", "error", err)
			http.Error(w, "Failed to convert endpoint to records", http.StatusInternalServerError)
			return
		}

		for _, record := range records {
			if err := client.UpdateDNSStaticRecord(record); err != nil {
				slog.Error("Failed to update record", "error", err)
				http.Error(w, "Failed to update record", http.StatusInternalServerError)
				return
			}
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func Healthz(w http.ResponseWriter, r *http.Request) {
	slog.Debug("GET /healthz", "request_headers", r.Header)

	w.WriteHeader(http.StatusOK)
}
