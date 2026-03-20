package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/afterdarksys/cariskscore/internal/ctsearch"
	"github.com/afterdarksys/cariskscore/internal/database"
	"github.com/afterdarksys/cariskscore/internal/indexer"
	"github.com/afterdarksys/cariskscore/internal/scorer"
)

const (
	defaultDBPath = "cariskscore.db"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Get database path
	dbPath := getDBPath()

	switch command {
	case "index":
		handleIndex(dbPath)
	case "score":
		handleScore(dbPath)
	case "list":
		handleList(dbPath)
	case "search":
		handleSearch(dbPath)
	case "top":
		handleTop(dbPath)
	case "stats":
		handleStats(dbPath)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func getDBPath() string {
	if dbPath := os.Getenv("CARISKSCORE_DB"); dbPath != "" {
		return dbPath
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return defaultDBPath
	}
	return filepath.Join(homeDir, ".cariskscore", defaultDBPath)
}

func ensureDBDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	return os.MkdirAll(dir, 0755)
}

func handleIndex(dbPath string) {
	if err := ensureDBDir(dbPath); err != nil {
		fmt.Printf("Error creating database directory: %v\n", err)
		os.Exit(1)
	}

	db, err := database.New(dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	idx := indexer.New(db)
	count, err := idx.IndexFromCCАDB()
	if err != nil {
		fmt.Printf("Error indexing CAs: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully indexed %d CAs\n", count)
}

func handleScore(dbPath string) {
	db, err := database.New(dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	scr := scorer.New(db)
	if err := scr.ComputeScores(); err != nil {
		fmt.Printf("Error computing scores: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Scores computed successfully")
}

func handleList(dbPath string) {
	db, err := database.New(dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	limit := 50
	if len(os.Args) > 2 {
		fmt.Sscanf(os.Args[2], "%d", &limit)
	}

	cas, err := db.ListCAs(limit, 0)
	if err != nil {
		fmt.Printf("Error listing CAs: %v\n", err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tOrganization\tCountry\tMozilla\tMicrosoft\tEV")
	for _, ca := range cas {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%v\t%v\t%v\n",
			ca.ID, ca.Name, ca.Organization, ca.Country,
			ca.TrustedByMozilla, ca.TrustedByMicrosoft, ca.EVCapable)
	}
	w.Flush()
}

func handleSearch(dbPath string) {
	if len(os.Args) < 3 {
		fmt.Println("Usage: cariskscore search <domain>")
		os.Exit(1)
	}

	domain := os.Args[2]
	cts := ctsearch.New()

	fmt.Printf("Searching Certificate Transparency logs for: %s\n", domain)
	results, err := cts.SearchByDomain(domain)
	if err != nil {
		fmt.Printf("Error searching CT logs: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Println("No certificates found")
		return
	}

	// Output as JSON for easier processing
	jsonOutput := os.Getenv("JSON_OUTPUT") != ""
	if jsonOutput {
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(data))
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tIssuer\tCommon Name\tNot Before\tNot After")
	for _, r := range results {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
			r.ID, r.IssuerName, r.CommonName,
			r.NotBefore.Format("2006-01-02"), r.NotAfter.Format("2006-01-02"))
	}
	w.Flush()

	fmt.Printf("\nTotal: %d certificates\n", len(results))
}

func handleTop(dbPath string) {
	db, err := database.New(dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	limit := 20
	if len(os.Args) > 2 {
		fmt.Sscanf(os.Args[2], "%d", &limit)
	}

	scr := scorer.New(db)
	scores, err := scr.GetTopCAs(limit)
	if err != nil {
		fmt.Printf("Error getting top CAs: %v\n", err)
		os.Exit(1)
	}

	if len(scores) == 0 {
		fmt.Println("No scores found. Run 'cariskscore score' first.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Rank\tCA ID\tOverall\tSecurity\tAudit\tIncident\tCompliance")
	for _, s := range scores {
		ca, err := db.GetCA(s.CAID)
		if err != nil {
			continue
		}
		fmt.Fprintf(w, "%d\t%s\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\n",
			s.Rank, ca.Name, s.OverallScore, s.SecurityScore,
			s.AuditScore, s.IncidentScore, s.ComplianceScore)
	}
	w.Flush()
}

func handleStats(dbPath string) {
	db, err := database.New(dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	count, err := db.CountCAs()
	if err != nil {
		fmt.Printf("Error getting stats: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Database: %s\n", dbPath)
	fmt.Printf("Total CAs indexed: %d\n", count)
}

func printUsage() {
	fmt.Println("CA Risk Score - Certificate Authority Security Assessment Tool")
	fmt.Println("\nUsage:")
	fmt.Println("  cariskscore <command> [arguments]")
	fmt.Println("\nCommands:")
	fmt.Println("  index          - Index CAs from CCADB")
	fmt.Println("  score          - Compute security scores for all CAs")
	fmt.Println("  list [N]       - List first N CAs (default: 50)")
	fmt.Println("  top [N]        - Show top N CAs by score (default: 20)")
	fmt.Println("  search <domain> - Search Certificate Transparency logs")
	fmt.Println("  stats          - Show database statistics")
	fmt.Println("\nEnvironment Variables:")
	fmt.Println("  CARISKSCORE_DB - Custom database path")
	fmt.Println("  JSON_OUTPUT    - Output search results as JSON")
}
