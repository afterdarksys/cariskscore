# CA Risk Score

A Go utility for indexing public Certificate Authorities, assessing their security risk, and searching Certificate Transparency logs.

## Features

1. **CA Indexing** - Fetches and indexes CAs from CCADB (Common CA Database)
2. **Risk Assessment** - Computes security scores based on trust status, audits, and incidents
3. **CA Scoring** - Ranks CAs from most to least secure
4. **Certificate Transparency Search** - Queries CT logs via crt.sh API

## Installation

```bash
go build -o cariskscore ./cmd/cariskscore
```

Or install directly:
```bash
go install ./cmd/cariskscore
```

## Usage

### Index CAs from CCADB

Download and index all public CAs from the Common CA Database:

```bash
./cariskscore index
```

This fetches CA data from CCADB's public CSV feed and stores it in a local SQLite database.

### Compute Security Scores

Calculate risk scores for all indexed CAs:

```bash
./cariskscore score
```

Scoring is based on:
- **Audit Score (25%)**: Trust by Mozilla, Microsoft, Chrome programs
- **Incident Score (30%)**: Number and severity of security incidents
- **Security Score (30%)**: Base security assessment
- **Compliance Score (15%)**: EV capability and multi-program trust

### View Top-Ranked CAs

Show the top 20 most secure CAs:

```bash
./cariskscore top
```

Show top 50:
```bash
./cariskscore top 50
```

### List CAs

List first 50 CAs in the database:

```bash
./cariskscore list
```

List first 100:
```bash
./cariskscore list 100
```

### Search Certificate Transparency Logs

Search for certificates issued for a domain:

```bash
./cariskscore search example.com
```

Output as JSON:
```bash
JSON_OUTPUT=1 ./cariskscore search example.com
```

### Database Statistics

View indexing statistics:

```bash
./cariskscore stats
```

## Configuration

### Environment Variables

- `CARISKSCORE_DB` - Custom database path (default: `~/.cariskscore/cariskscore.db`)
- `JSON_OUTPUT` - Output search results as JSON (for `search` command)

### Database Location

By default, the database is stored at `~/.cariskscore/cariskscore.db`. Change this with:

```bash
export CARISKSCORE_DB=/path/to/custom/db.sqlite
```

## Architecture

### Components

- **Indexer** (`internal/indexer`) - Fetches CA data from CCADB
- **Scorer** (`internal/scorer`) - Computes risk scores for CAs
- **CT Searcher** (`internal/ctsearch`) - Queries Certificate Transparency logs
- **Database** (`internal/database`) - SQLite storage for CAs, risks, and scores
- **Models** (`internal/models`) - Data structures

### Data Sources

- **CCADB**: https://ccadb.my.salesforce-sites.com - Common CA Database
- **crt.sh**: https://crt.sh - Certificate Transparency log aggregator

## Example Workflow

```bash
# 1. Index CAs
./cariskscore index
# Output: Successfully indexed 500+ CAs

# 2. Compute scores
./cariskscore score
# Output: Computing scores for 500 CAs...
#         Score computation complete

# 3. View top CAs
./cariskscore top 10

# 4. Search CT logs
./cariskscore search github.com
```

## Scoring Algorithm

Each CA receives scores in four categories (0-100, higher is better):

1. **Security Score**: Base security assessment (currently 100 for all)
2. **Audit Score**: Based on trust by major root programs
   - Mozilla: +35 points
   - Microsoft: +35 points
   - Chrome: +30 points
3. **Incident Score**: Deductions for security incidents
   - Critical: -30 points
   - High: -15 points
   - Medium: -8 points
   - Low: -3 points
4. **Compliance Score**: Based on trust breadth and EV capability

**Overall Score** = Weighted average:
- Security: 30%
- Audit: 25%
- Incident: 30%
- Compliance: 15%

## Development

### Project Structure

```
cariskscore/
├── cmd/
│   └── cariskscore/      # CLI application
├── internal/
│   ├── database/         # SQLite database layer
│   ├── indexer/          # CA indexing from CCADB
│   ├── scorer/           # Risk scoring engine
│   ├── ctsearch/         # CT log search
│   └── models/           # Data models
├── go.mod
└── README.md
```

### Dependencies

- `github.com/mattn/go-sqlite3` - SQLite driver

### Building

```bash
go build -o cariskscore ./cmd/cariskscore
```

### Testing

```bash
go test ./...
```

## Future Enhancements

- [ ] Import security incident data from CVE databases
- [ ] Track CT log violations
- [ ] Add audit report parsing
- [ ] Web UI for visualization
- [ ] API server mode
- [ ] Export reports (PDF, CSV)
- [ ] Integration with Censys API
- [ ] Historical trend tracking

## License

MIT

## Contributing

Contributions welcome! Please open an issue or PR.
