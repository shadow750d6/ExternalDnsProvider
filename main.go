package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/libdns/duckdns"
	"github.com/libdns/libdns"
	myaddr_dns_provider "github.com/shadow750d6/myaddr-dns-provider/lib"
)

type CombinedDnsProvider struct {
	Myaddr  *myaddr_dns_provider.Provider
	Duckdns *duckdns.Provider
}

// AppendRecords adds records to a zone. It returns the records that were added.
func (p *CombinedDnsProvider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	switch strings.Trim(zone, ".") {
	case "duckdns.org":
		return p.Duckdns.AppendRecords(ctx, zone, records)
	case "myaddr.io":
	case "myaddr.dev":
	case "myaddr.tools":
		return p.Myaddr.AppendRecords(ctx, zone, records)
	}
	return nil, fmt.Errorf("unsupported zone %s", zone)
}

// DeleteRecords deletes records from a zone. It returns the records that were deleted.
func (p *CombinedDnsProvider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	switch strings.Trim(zone, ".") {
	case "duckdns.org":
		return p.Duckdns.DeleteRecords(ctx, zone, records)
	case "myaddr.io":
	case "myaddr.dev":
	case "myaddr.tools":
		return p.Myaddr.DeleteRecords(ctx, zone, records)
	}
	return nil, fmt.Errorf("unsupported zone %s", zone)
}

type ExternalDnsProviderRequest struct {
	Method string
	Zone   string
	RRs    []libdns.RR
}

type ExternalDnsProviderResponse struct {
	Error string
	RRs   []libdns.RR
}

func main() {
	// Parse environment variables
	requestJSON := os.Getenv("LIBDNS_REQUEST")
	providerJSON := os.Getenv("COMBINED_DNS_PROVIDER")

	var request ExternalDnsProviderRequest
	var provider CombinedDnsProvider

	if err := json.Unmarshal([]byte(requestJSON), &request); err != nil {
		log.Fatalf("Error unmarshaling LIBDNS_REQUEST: %v", err)
	}

	if err := json.Unmarshal([]byte(providerJSON), &provider); err != nil {
		log.Fatalf("Error unmarshaling COMBINED_DNS_PROVIDER: %v", err)
	}

	// Process the request
	response := ExternalDnsProviderResponse{}

	inputRecords, err := convertRRtoRecord(request.RRs)
	if err != nil {
		response.Error = err.Error()
	} else {
		switch request.Method {
		case "append":
			records, err := provider.AppendRecords(context.Background(), request.Zone, inputRecords)
			if err != nil {
				response.Error = err.Error()
			}
			response.RRs = convertRecordtoRR(records)
		case "delete":
			records, err := provider.DeleteRecords(context.Background(), request.Zone, inputRecords)
			if err != nil {
				response.Error = err.Error()
			}
			response.RRs = convertRecordtoRR(records)
		default:
			response.Error = fmt.Sprintf("unsupported method %s", request.Method)
		}
	}
	// Write the response as JSON
	responseJSON, err := json.Marshal(response)
	if err != nil {
		log.Fatalf("Error marshaling response: %v", err)
	}

	fmt.Println(string(responseJSON))
}

func convertRRtoRecord(rrs []libdns.RR) ([]libdns.Record, error) {
	records := make([]libdns.Record, len(rrs))
	for i, rr := range rrs {
		record, err := rr.Parse()
		if err != nil {
			return nil, err
		}
		records[i] = record
	}
	return records, nil
}

func convertRecordtoRR(records []libdns.Record) []libdns.RR {
	rrs := make([]libdns.RR, len(records))
	for i, record := range records {
		rrs[i] = record.RR()
	}
	return rrs
}

func (r ExternalDnsProviderRequest) Validate() error {
	if r.Method == "" {
		return errors.New("Method is required")
	}
	return nil
}
