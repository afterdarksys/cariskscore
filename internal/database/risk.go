package database

import (
	"fmt"
	"time"

	"github.com/afterdarksys/cariskscore/internal/models"
)

// InsertRiskFactor inserts a new risk factor
func (db *DB) InsertRiskFactor(rf *models.RiskFactor) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	stmt, err := db.conn.Prepare(`
		INSERT INTO risk_factors (ca_id, risk_type, severity, description, source, detected_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(rf.CAID, rf.RiskType, rf.Severity, rf.Description, rf.Source, rf.DetectedAt)
	if err != nil {
		return 0, fmt.Errorf("failed to insert risk factor: %w", err)
	}

	id, err := result.LastInsertId()
	return int(id), err
}

// GetRiskFactorsForCA retrieves all risk factors for a CA
func (db *DB) GetRiskFactorsForCA(caID int) ([]*models.RiskFactor, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	query := `
		SELECT id, ca_id, risk_type, severity, description, source, detected_at, resolved_at
		FROM risk_factors
		WHERE ca_id = ? AND resolved_at IS NULL
		ORDER BY severity, detected_at DESC
	`

	rows, err := db.conn.Query(query, caID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rfs []*models.RiskFactor
	for rows.Next() {
		rf := &models.RiskFactor{}
		if err := rows.Scan(
			&rf.ID, &rf.CAID, &rf.RiskType, &rf.Severity, &rf.Description, &rf.Source, &rf.DetectedAt, &rf.ResolvedAt,
		); err != nil {
			return nil, err
		}
		rfs = append(rfs, rf)
	}

	return rfs, rows.Err()
}

// UpsertCAScore inserts or updates a CA score
func (db *DB) UpsertCAScore(score *models.CAScore) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	stmt, err := db.conn.Prepare(`
		INSERT INTO ca_scores (ca_id, security_score, audit_score, incident_score, compliance_score, overall_score, rank, computed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(ca_id) DO UPDATE SET
			security_score = excluded.security_score,
			audit_score = excluded.audit_score,
			incident_score = excluded.incident_score,
			compliance_score = excluded.compliance_score,
			overall_score = excluded.overall_score,
			rank = excluded.rank,
			computed_at = excluded.computed_at
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		score.CAID, score.SecurityScore, score.AuditScore, score.IncidentScore,
		score.ComplianceScore, score.OverallScore, score.Rank, time.Now(),
	)
	return err
}

// GetCAScore retrieves the score for a CA
func (db *DB) GetCAScore(caID int) (*models.CAScore, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	score := &models.CAScore{}
	err := db.conn.QueryRow(`
		SELECT id, ca_id, security_score, audit_score, incident_score, compliance_score, overall_score, rank, computed_at
		FROM ca_scores WHERE ca_id = ?
	`, caID).Scan(
		&score.ID, &score.CAID, &score.SecurityScore, &score.AuditScore, &score.IncidentScore,
		&score.ComplianceScore, &score.OverallScore, &score.Rank, &score.ComputedAt,
	)
	if err != nil {
		return nil, err
	}
	return score, nil
}

// ListScoresSorted retrieves all CA scores sorted by overall score (descending)
func (db *DB) ListScoresSorted(limit int, offset int) ([]*models.CAScore, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	query := `
		SELECT s.id, s.ca_id, s.security_score, s.audit_score, s.incident_score, s.compliance_score, s.overall_score, s.rank, s.computed_at
		FROM ca_scores s
		ORDER BY s.overall_score DESC
		LIMIT ? OFFSET ?
	`

	rows, err := db.conn.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []*models.CAScore
	for rows.Next() {
		score := &models.CAScore{}
		if err := rows.Scan(
			&score.ID, &score.CAID, &score.SecurityScore, &score.AuditScore, &score.IncidentScore,
			&score.ComplianceScore, &score.OverallScore, &score.Rank, &score.ComputedAt,
		); err != nil {
			return nil, err
		}
		scores = append(scores, score)
	}

	return scores, rows.Err()
}
