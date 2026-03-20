package models

import "time"

// CertificateAuthority represents a CA with its metadata
type CertificateAuthority struct {
	ID                 int       `json:"id"`
	Name               string    `json:"name"`
	CommonName         string    `json:"common_name"`
	Organization       string    `json:"organization"`
	Country            string    `json:"country"`
	CertificateSHA256  string    `json:"certificate_sha256"`
	Source             string    `json:"source"` // Mozilla, Microsoft, CCADB
	TrustedByMozilla   bool      `json:"trusted_by_mozilla"`
	TrustedByMicrosoft bool      `json:"trusted_by_microsoft"`
	TrustedByChrome    bool      `json:"trusted_by_chrome"`
	EVCapable          bool      `json:"ev_capable"`
	IndexedAt          time.Time `json:"indexed_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// RiskFactor represents a security risk associated with a CA
type RiskFactor struct {
	ID          int        `json:"id"`
	CAID        int        `json:"ca_id"`
	RiskType    string     `json:"risk_type"` // incident, audit_failure, ct_violation, etc.
	Severity    string     `json:"severity"`  // critical, high, medium, low
	Description string     `json:"description"`
	Source      string     `json:"source"`
	DetectedAt  time.Time  `json:"detected_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// CAScore represents the computed risk score for a CA
type CAScore struct {
	ID              int       `json:"id"`
	CAID            int       `json:"ca_id"`
	SecurityScore   float64   `json:"security_score"`   // 0-100, higher is better
	AuditScore      float64   `json:"audit_score"`      // 0-100
	IncidentScore   float64   `json:"incident_score"`   // 0-100
	ComplianceScore float64   `json:"compliance_score"` // 0-100
	OverallScore    float64   `json:"overall_score"`    // Weighted average
	Rank            int       `json:"rank"`
	ComputedAt      time.Time `json:"computed_at"`
}

// CTLogEntry represents a certificate from Certificate Transparency logs
type CTLogEntry struct {
	ID             int64     `json:"id"`
	SerialNumber   string    `json:"serial_number"`
	IssuerCAName   string    `json:"issuer_ca_name"`
	CommonName     string    `json:"common_name"`
	NotBefore      time.Time `json:"not_before"`
	NotAfter       time.Time `json:"not_after"`
	LoggedAt       time.Time `json:"logged_at"`
	CertificatePEM string    `json:"certificate_pem"`
}
