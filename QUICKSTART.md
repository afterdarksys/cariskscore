# Quick Start Guide

## Build

```bash
go build -o cariskscore ./cmd/cariskscore
```

## Basic Usage

### 1. Index CAs (required first step)

```bash
./cariskscore index
```

This downloads and indexes 500+ public CAs from CCADB. Takes ~30 seconds.

### 2. Compute Scores

```bash
./cariskscore score
```

Calculates security scores for all indexed CAs.

### 3. View Results

```bash
# Show top 20 most secure CAs
./cariskscore top

# Show top 50
./cariskscore top 50

# List all CAs
./cariskscore list

# Database stats
./cariskscore stats
```

### 4. Search Certificate Transparency Logs

```bash
# Search for a domain
./cariskscore search github.com

# Output as JSON
JSON_OUTPUT=1 ./cariskscore search example.com
```

## Examples

### Full Workflow

```bash
# Build
go build -o cariskscore ./cmd/cariskscore

# Index CAs from CCADB
./cariskscore index
# Output: Successfully indexed 523 CAs

# Compute security scores
./cariskscore score
# Output: Computing scores for 523 CAs...
#         Score computation complete

# View top 10 most secure CAs
./cariskscore top 10

# Search CT logs
./cariskscore search google.com
```

### Output Examples

**Top CAs:**
```
Rank  CA ID                      Overall  Security  Audit  Incident  Compliance
1     DigiCert                   95.00    100.00    100.00 100.00    90.00
2     Let's Encrypt              94.50    100.00    100.00 100.00    85.00
3     GlobalSign                 94.00    100.00    100.00 100.00    85.00
```

**CT Search:**
```
ID        Issuer                        Common Name       Not Before  Not After
12345678  Let's Encrypt Authority X3    example.com       2024-01-15  2024-04-15
12345679  DigiCert SHA2 Secure Server   *.example.com     2024-02-01  2025-02-01
```

## Configuration

### Custom Database Location

```bash
export CARISKSCORE_DB=/custom/path/cariskscore.db
./cariskscore index
```

Default: `~/.cariskscore/cariskscore.db`

## Data Sources

- **CCADB**: Common CA Database with 500+ trusted CAs
- **crt.sh**: Certificate Transparency log aggregator

## What Gets Scored?

Each CA receives 4 scores (0-100, higher = better):

1. **Security Score** (30% weight): Base security assessment
2. **Audit Score** (25% weight): Trust by Mozilla/Microsoft/Chrome
3. **Incident Score** (30% weight): Security incident history
4. **Compliance Score** (15% weight): EV capability and trust breadth

## Troubleshooting

### Build fails with SQLite error
Make sure you have gcc installed:
```bash
# macOS
xcode-select --install

# Linux
sudo apt-get install gcc
```

### Database locked
Close any other processes accessing the database.

### No CAs found
Run `./cariskscore index` first to populate the database.

## Next Steps

- See [README.md](README.md) for full documentation
- Check the [Architecture](README.md#architecture) section for implementation details
- Review [Future Enhancements](README.md#future-enhancements) for potential contributions
