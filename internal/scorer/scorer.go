package scorer

import (
	"fmt"
	"math"

	"github.com/afterdarksys/cariskscore/internal/database"
	"github.com/afterdarksys/cariskscore/internal/models"
)

type Scorer struct {
	db *database.DB
}

func New(db *database.DB) *Scorer {
	return &Scorer{db: db}
}

// ComputeScores calculates risk scores for all CAs
func (s *Scorer) ComputeScores() error {
	// Get all CAs
	cas, err := s.db.ListCAs(10000, 0) // Get all CAs
	if err != nil {
		return fmt.Errorf("failed to list CAs: %w", err)
	}

	fmt.Printf("Computing scores for %d CAs...\n", len(cas))

	// Compute score for each CA
	for _, ca := range cas {
		score, err := s.computeCAScore(ca)
		if err != nil {
			fmt.Printf("Warning: failed to compute score for %s: %v\n", ca.Name, err)
			continue
		}

		if err := s.db.UpsertCAScore(score); err != nil {
			fmt.Printf("Warning: failed to save score for %s: %v\n", ca.Name, err)
		}
	}

	// Update ranks
	if err := s.updateRanks(); err != nil {
		return fmt.Errorf("failed to update ranks: %w", err)
	}

	fmt.Println("Score computation complete")
	return nil
}

func (s *Scorer) computeCAScore(ca *models.CertificateAuthority) (*models.CAScore, error) {
	// Base security score starts at 100
	securityScore := 100.0

	// Audit score based on trust status
	auditScore := s.calculateAuditScore(ca)

	// Get risk factors for this CA
	riskFactors, err := s.db.GetRiskFactorsForCA(ca.ID)
	if err != nil {
		return nil, err
	}

	// Incident score based on risk factors
	incidentScore := s.calculateIncidentScore(riskFactors)

	// Compliance score based on trust bits and EV capability
	complianceScore := s.calculateComplianceScore(ca)

	// Overall weighted score
	// Weights: Security 30%, Audit 25%, Incident 30%, Compliance 15%
	overallScore := (securityScore * 0.30) +
		(auditScore * 0.25) +
		(incidentScore * 0.30) +
		(complianceScore * 0.15)

	return &models.CAScore{
		CAID:            ca.ID,
		SecurityScore:   math.Round(securityScore*100) / 100,
		AuditScore:      math.Round(auditScore*100) / 100,
		IncidentScore:   math.Round(incidentScore*100) / 100,
		ComplianceScore: math.Round(complianceScore*100) / 100,
		OverallScore:    math.Round(overallScore*100) / 100,
	}, nil
}

func (s *Scorer) calculateAuditScore(ca *models.CertificateAuthority) float64 {
	score := 0.0

	// Trust by major programs indicates passing audits
	if ca.TrustedByMozilla {
		score += 35.0
	}
	if ca.TrustedByMicrosoft {
		score += 35.0
	}
	if ca.TrustedByChrome {
		score += 30.0
	}

	// If not trusted by any, give base score
	if score == 0 {
		score = 50.0 // Neutral score for unknown CAs
	}

	return math.Min(score, 100.0)
}

func (s *Scorer) calculateIncidentScore(risks []*models.RiskFactor) float64 {
	if len(risks) == 0 {
		return 100.0 // No incidents = perfect score
	}

	// Deduct points based on severity
	deductions := 0.0
	for _, risk := range risks {
		switch risk.Severity {
		case "critical":
			deductions += 30.0
		case "high":
			deductions += 15.0
		case "medium":
			deductions += 8.0
		case "low":
			deductions += 3.0
		}
	}

	score := 100.0 - deductions
	return math.Max(score, 0.0)
}

func (s *Scorer) calculateComplianceScore(ca *models.CertificateAuthority) float64 {
	score := 50.0 // Base score

	// Trusted by multiple programs indicates good compliance
	trustCount := 0
	if ca.TrustedByMozilla {
		trustCount++
		score += 15.0
	}
	if ca.TrustedByMicrosoft {
		trustCount++
		score += 15.0
	}
	if ca.TrustedByChrome {
		trustCount++
		score += 10.0
	}

	// EV capability shows advanced compliance
	if ca.EVCapable {
		score += 10.0
	}

	return math.Min(score, 100.0)
}

func (s *Scorer) updateRanks() error {
	// Get all scores ordered by overall score
	scores, err := s.db.ListScoresSorted(10000, 0)
	if err != nil {
		return err
	}

	// Update ranks
	for i, score := range scores {
		score.Rank = i + 1
		if err := s.db.UpsertCAScore(score); err != nil {
			return err
		}
	}

	return nil
}

// GetTopCAs returns the top N CAs by score
func (s *Scorer) GetTopCAs(limit int) ([]*models.CAScore, error) {
	return s.db.ListScoresSorted(limit, 0)
}
