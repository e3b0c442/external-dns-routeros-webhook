package main

import (
	"fmt"
	"log/slog"

	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/provider"
)

func recordToEndpoint(record *Record) (*endpoint.Endpoint, error) {
	var target string
	switch record.Type {
	case "A":
		fallthrough
	case "AAAA":
		target = record.Address
	case "CNAME":
		target = record.CName
	case "NS":
		target = record.NS
	case "SRV":
		target = fmt.Sprintf("%s %s %s %s", record.SrvPriority, record.SrvWeight, record.SrvPort, record.SrvTarget)
	case "TXT":
		target = record.Text
	default:
		return nil, ErrUnsupportedRecordType
	}

	return &endpoint.Endpoint{
		DNSName:    record.Name,
		Targets:    endpoint.Targets{target},
		RecordType: record.Type,
		RecordTTL:  endpoint.TTL(record.TTL),
		ProviderSpecific: endpoint.ProviderSpecific{
			{Name: ".id", Value: record.ID},
		},
	}, nil
}

func recordsToEndpoints(records []*Record) ([]*endpoint.Endpoint, error) {
	epm := make(map[string]map[string]*endpoint.Endpoint)

	for _, record := range records {
		if !provider.SupportedRecordType(record.Type) {
			slog.Warn("Unsupported record type", "name", record.Name, "type", record.Type)
			continue
		}
		ep, err := recordToEndpoint(record)
		if err != nil {
			return nil, err
		}

		if _, ok := epm[record.Name]; !ok {
			epm[record.Name] = make(map[string]*endpoint.Endpoint)
		}
		if _, ok := epm[record.Name][record.Type]; !ok {
			epm[record.Name][record.Type] = ep
		} else {
			epm[record.Name][record.Type].Targets = append(epm[record.Name][record.Type].Targets, ep.Targets...)
		}
	}

	var eps []*endpoint.Endpoint
	for _, gp := range epm {
		for _, ep := range gp {
			eps = append(eps, ep)
		}
	}

	return eps, nil
}

func endpointToRecords(ep *endpoint.Endpoint) ([]*Record, error) {
	var records []*Record

	for _, target := range ep.Targets {
		var record Record
		record.Name = ep.DNSName
		record.Type = ep.RecordType
		record.TTL = TTL(ep.RecordTTL)

		switch ep.RecordType {
		case "A":
			fallthrough
		case "AAAA":
			record.Address = ep.Targets[0]
		case "CNAME":
			record.CName = ep.Targets[0]
		case "NS":
			record.NS = ep.Targets[0]
		case "SRV":
			var err error
			n, err := fmt.Sscanf(target, "%s %s %s %s", &record.SrvPriority, &record.SrvWeight, &record.SrvPort, &record.SrvTarget)
			if err != nil {
				return nil, err
			}
			if n != 4 {
				return nil, fmt.Errorf("failed to parse SRV target")
			}
		case "TXT":
			record.Text = ep.Targets[0]
		default:
			slog.Warn("Unsupported record type", "name", ep.DNSName, "type", ep.RecordType)
		}

		for _, ps := range ep.ProviderSpecific {
			switch ps.Name {
			case ".id":
				record.ID = ps.Value
			}
		}

		records = append(records, &record)
	}

	return records, nil
}

func endpointsToRecords(eps []*endpoint.Endpoint) ([]*Record, error) {
	var records []*Record

	for _, ep := range eps {
		rs, err := endpointToRecords(ep)
		if err != nil {
			return nil, err
		}
		records = append(records, rs...)
	}

	return records, nil
}
