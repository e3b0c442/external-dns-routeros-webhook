package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type RouterOSAPIError struct {
	Error   int    `json:"error"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

type RouterOSAPIClient struct {
	username, password, baseURL string
	*http.Client
}

func NewRouterOSAPIClient(username, password, baseURL string) *RouterOSAPIClient {
	return &RouterOSAPIClient{
		username: username,
		password: password,
		baseURL:  baseURL,
		Client:   &http.Client{},
	}
}

func (c *RouterOSAPIClient) ReadDNSStaticRecords() ([]*Record, error) {
	var records []*Record

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, c.baseURL+"/rest/ip/dns/static", nil)
	if err != nil {
		return nil, fmt.Errorf("ListDNSStaticRecords: failed to create request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)
	slog.Debug("ListDNSStaticRecords", "url", req.URL.String())

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ListDNSStaticRecords: failed to get records: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var err RouterOSAPIError
		if err := json.NewDecoder(resp.Body).Decode(&err); err != nil {
			return nil, fmt.Errorf("ListDNSStaticRecords: failed to decode error: %w", err)
		}
		return nil, fmt.Errorf("ListDNSStaticRecords: failed to get records: %d %s %s", err.Error, err.Message, err.Detail)
	}

	if err := json.NewDecoder(resp.Body).Decode(&records); err != nil {
		return nil, fmt.Errorf("ListDNSStaticRecords: failed to decode records: %w", err)
	}

	return records, nil
}

func (c *RouterOSAPIClient) CreateDNSStaticRecord(record *Record) error {
	body := &bytes.Buffer{}
	json.NewEncoder(body).Encode(record)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, c.baseURL+"/rest/ip/dns/static", body)
	if err != nil {
		return fmt.Errorf("CreateDNSStaticRecord: failed to create request: %w", err)
	}
	slog.Debug("CreateDNSStaticRecord", "url", req.URL.String())

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("CreateDNSStaticRecord: failed to create record: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		var err RouterOSAPIError
		if err := json.NewDecoder(resp.Body).Decode(&err); err != nil {
			return fmt.Errorf("CreateDNSStaticRecord: failed to decode error: %w", err)
		}
		return fmt.Errorf("CreateDNSStaticRecord: failed to create record: %d %s %s", err.Error, err.Message, err.Detail)
	}

	return nil
}

func (c *RouterOSAPIClient) DeleteDNSStaticRecord(id string) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, c.baseURL+"/rest/ip/dns/static/"+id, nil)
	if err != nil {
		return fmt.Errorf("DeleteDNSStaticRecord: failed to create request: %w", err)
	}
	slog.Debug("DeleteDNSStaticRecord", "url", req.URL.String())
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("DeleteDNSStaticRecord: failed to delete record: %w", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		var err RouterOSAPIError
		if err := json.NewDecoder(resp.Body).Decode(&err); err != nil {
			return fmt.Errorf("DeleteDNSStaticRecord: failed to decode error: %w", err)
		}
		return fmt.Errorf("DeleteDNSStaticRecord: failed to delete record: %d %s %s", err.Error, err.Message, err.Detail)
	}

	return nil
}

func (c *RouterOSAPIClient) UpdateDNSStaticRecord(record *Record) error {
	body := &bytes.Buffer{}
	json.NewEncoder(body).Encode(record)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPatch, c.baseURL+"/rest/ip/dns/static/"+record.ID, body)
	if err != nil {
		return fmt.Errorf("UpdateDNSStaticRecord: failed to create request: %w", err)
	}
	slog.Debug("UpdateDNSStaticRecord", "url", req.URL.String())

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("UpdateDNSStaticRecord: failed to update record: %w", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		var err RouterOSAPIError
		if err := json.NewDecoder(resp.Body).Decode(&err); err != nil {
			return fmt.Errorf("UpdateDNSStaticRecord: failed to decode error: %w", err)
		}
		return fmt.Errorf("UpdateDNSStaticRecord: failed to update record: %d %s %s", err.Error, err.Message, err.Detail)
	}

	return nil
}
