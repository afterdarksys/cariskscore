package ctsearch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const crtshAPIURL = "https://crt.sh"

type CTSearcher struct {
	client *http.Client
}

// CTResult represents a certificate from crt.sh
type CTResult struct {
	IssuerCAID  int       `json:"issuer_ca_id"`
	IssuerName  string    `json:"issuer_name"`
	CommonName  string    `json:"common_name"`
	NameValue   string    `json:"name_value"`
	ID          int64     `json:"id"`
	EntryID     int64     `json:"entry_id"`
	NotBefore   time.Time `json:"not_before"`
	NotAfter    time.Time `json:"not_after"`
	SerialNumber string   `json:"serial_number"`
}

func New() *CTSearcher {
	return &CTSearcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchByDomain searches CT logs for certificates matching a domain
func (cts *CTSearcher) SearchByDomain(domain string) ([]*CTResult, error) {
	return cts.search(map[string]string{
		"q": domain,
		"output": "json",
	})
}

// SearchByCAName searches CT logs for certificates issued by a specific CA
func (cts *CTSearcher) SearchByCAName(caName string) ([]*CTResult, error) {
	return cts.search(map[string]string{
		"Identity": caName,
		"output": "json",
	})
}

// SearchByIssuerID searches by crt.sh CA ID
func (cts *CTSearcher) SearchByIssuerID(issuerID int) ([]*CTResult, error) {
	return cts.search(map[string]string{
		"caid": fmt.Sprintf("%d", issuerID),
		"output": "json",
	})
}

func (cts *CTSearcher) search(params map[string]string) ([]*CTResult, error) {
	// Build query URL
	queryURL, err := url.Parse(crtshAPIURL)
	if err != nil {
		return nil, err
	}

	q := queryURL.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	queryURL.RawQuery = q.Encode()

	// Make request
	resp, err := cts.client.Get(queryURL.String())
	if err != nil {
		return nil, fmt.Errorf("crt.sh request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("crt.sh returned status %d", resp.StatusCode)
	}

	// Parse JSON response
	var results []*CTResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to parse crt.sh response: %w", err)
	}

	return results, nil
}

// GetCertificateByID fetches detailed certificate info by crt.sh ID
func (cts *CTSearcher) GetCertificateByID(id int64) (string, error) {
	certURL := fmt.Sprintf("%s/?id=%d&output=json", crtshAPIURL, id)
	
	resp, err := cts.client.Get(certURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch certificate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("crt.sh returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse certificate: %w", err)
	}

	if pem, ok := result["pem"].(string); ok {
		return pem, nil
	}

	return "", fmt.Errorf("no PEM data found")
}
