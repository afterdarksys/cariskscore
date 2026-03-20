package indexer

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/afterdarksys/cariskscore/internal/database"
	"github.com/afterdarksys/cariskscore/internal/models"
)

const (
	// CCADB All Certificate Information CSV
	ccadbAllCertsURL = "https://ccadb.my.salesforce-sites.com/ccadb/AllCertificateRecordsCSVFormatv2"
)

type Indexer struct {
	db     *database.DB
	client *http.Client
}

func New(db *database.DB) *Indexer {
	return &Indexer{
		db: db,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IndexFromCCАDB fetches and indexes CAs from the CCADB
func (idx *Indexer) IndexFromCCАDB() (int, error) {
	fmt.Println("Fetching CA data from CCADB...")
	resp, err := idx.client.Get(ccadbAllCertsURL)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch CCADB data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("CCADB returned status %d", resp.StatusCode)
	}

	reader := csv.NewReader(resp.Body)
	headers, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	// Map header names to column indices
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[h] = i
	}

	indexed := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return indexed, fmt.Errorf("error reading CSV: %w", err)
		}

		ca := idx.parseCAFromCCАDB(record, headerMap)
		if ca == nil {
			continue
		}

		// Check if CA already exists
		exists, err := idx.db.CAExists(ca.Name)
		if err != nil {
			return indexed, err
		}

		if !exists {
			_, err := idx.db.InsertCA(ca)
			if err != nil {
				fmt.Printf("Warning: failed to insert CA %s: %v\n", ca.Name, err)
				continue
			}
			indexed++
		}
	}

	fmt.Printf("Indexed %d CAs from CCADB\n", indexed)
	return indexed, nil
}

func (idx *Indexer) parseCAFromCCАDB(record []string, headerMap map[string]int) *models.CertificateAuthority {
	// Helper to safely get column value
	getCol := func(name string) string {
		if i, ok := headerMap[name]; ok && i < len(record) {
			return strings.TrimSpace(record[i])
		}
		return ""
	}

	caOwner := getCol("CA Owner")
	certName := getCol("Certificate Name")
	if caOwner == "" && certName == "" {
		return nil
	}

	name := caOwner
	if name == "" {
		name = certName
	}

	// Check trust bits
	mozillaStatus := strings.ToLower(getCol("Mozilla Status"))
	microsoftStatus := strings.ToLower(getCol("Microsoft Status"))
	
	trustedByMozilla := strings.Contains(mozillaStatus, "included")
	trustedByMicrosoft := strings.Contains(microsoftStatus, "included")

	// EV capable check
	evPolicy := getCol("EV Policy OID(s)")
	evCapable := evPolicy != ""

	ca := &models.CertificateAuthority{
		Name:               name,
		CommonName:         certName,
		Organization:       caOwner,
		Country:            getCol("Country"),
		CertificateSHA256:  getCol("SHA-256 Fingerprint"),
		Source:             "CCADB",
		TrustedByMozilla:   trustedByMozilla,
		TrustedByMicrosoft: trustedByMicrosoft,
		TrustedByChrome:    false, // CCADB doesn't explicitly track Chrome, but Chrome uses CCADB
		EVCapable:          evCapable,
	}

	return ca
}

// IndexFromMozilla fetches CAs from Mozilla's included CA list (via CCADB API)
func (idx *Indexer) IndexFromMozilla() (int, error) {
	// Mozilla's CA list is part of CCADB, so we use the same source
	// This is a convenience method
	fmt.Println("Mozilla CA data is included in CCADB index")
	return 0, nil
}

// Stats returns indexing statistics
func (idx *Indexer) Stats() (map[string]int, error) {
	total, err := idx.db.CountCAs()
	if err != nil {
		return nil, err
	}

	stats := map[string]int{
		"total": total,
	}

	return stats, nil
}
