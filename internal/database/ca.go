package database

import (
	"fmt"
	"time"

	"github.com/afterdarksys/cariskscore/internal/models"
)

// InsertCA inserts a new CA into the database
func (db *DB) InsertCA(ca *models.CertificateAuthority) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	stmt, err := db.conn.Prepare(`
		INSERT INTO certificate_authorities 
		(name, common_name, organization, country, certificate_sha256, source, 
		 trusted_by_mozilla, trusted_by_microsoft, trusted_by_chrome, ev_capable)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(
		ca.Name, ca.CommonName, ca.Organization, ca.Country, ca.CertificateSHA256,
		ca.Source, ca.TrustedByMozilla, ca.TrustedByMicrosoft, ca.TrustedByChrome, ca.EVCapable,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert CA: %w", err)
	}

	id, err := result.LastInsertId()
	return int(id), err
}

// GetCA retrieves a CA by ID
func (db *DB) GetCA(id int) (*models.CertificateAuthority, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	ca := &models.CertificateAuthority{}
	err := db.conn.QueryRow(`
		SELECT id, name, common_name, organization, country, certificate_sha256,
		       source, trusted_by_mozilla, trusted_by_microsoft, trusted_by_chrome,
		       ev_capable, indexed_at, updated_at
		FROM certificate_authorities WHERE id = ?
	`, id).Scan(
		&ca.ID, &ca.Name, &ca.CommonName, &ca.Organization, &ca.Country, &ca.CertificateSHA256,
		&ca.Source, &ca.TrustedByMozilla, &ca.TrustedByMicrosoft, &ca.TrustedByChrome,
		&ca.EVCapable, &ca.IndexedAt, &ca.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return ca, nil
}

// GetCAByName retrieves a CA by name
func (db *DB) GetCAByName(name string) (*models.CertificateAuthority, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	ca := &models.CertificateAuthority{}
	err := db.conn.QueryRow(`
		SELECT id, name, common_name, organization, country, certificate_sha256,
		       source, trusted_by_mozilla, trusted_by_microsoft, trusted_by_chrome,
		       ev_capable, indexed_at, updated_at
		FROM certificate_authorities WHERE name = ?
	`, name).Scan(
		&ca.ID, &ca.Name, &ca.CommonName, &ca.Organization, &ca.Country, &ca.CertificateSHA256,
		&ca.Source, &ca.TrustedByMozilla, &ca.TrustedByMicrosoft, &ca.TrustedByChrome,
		&ca.EVCapable, &ca.IndexedAt, &ca.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return ca, nil
}

// ListCAs retrieves all CAs, optionally filtered
func (db *DB) ListCAs(limit int, offset int) ([]*models.CertificateAuthority, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	query := `
		SELECT id, name, common_name, organization, country, certificate_sha256,
		       source, trusted_by_mozilla, trusted_by_microsoft, trusted_by_chrome,
		       ev_capable, indexed_at, updated_at
		FROM certificate_authorities
		ORDER BY name
		LIMIT ? OFFSET ?
	`

	rows, err := db.conn.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cas []*models.CertificateAuthority
	for rows.Next() {
		ca := &models.CertificateAuthority{}
		if err := rows.Scan(
			&ca.ID, &ca.Name, &ca.CommonName, &ca.Organization, &ca.Country, &ca.CertificateSHA256,
			&ca.Source, &ca.TrustedByMozilla, &ca.TrustedByMicrosoft, &ca.TrustedByChrome,
			&ca.EVCapable, &ca.IndexedAt, &ca.UpdatedAt,
		); err != nil {
			return nil, err
		}
		cas = append(cas, ca)
	}

	return cas, rows.Err()
}

// CountCAs returns the total number of CAs in the database
func (db *DB) CountCAs() (int, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM certificate_authorities").Scan(&count)
	return count, err
}

// UpdateCA updates an existing CA
func (db *DB) UpdateCA(ca *models.CertificateAuthority) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	stmt, err := db.conn.Prepare(`
		UPDATE certificate_authorities 
		SET common_name = ?, organization = ?, country = ?, certificate_sha256 = ?,
		    trusted_by_mozilla = ?, trusted_by_microsoft = ?, trusted_by_chrome = ?,
		    ev_capable = ?, updated_at = ?
		WHERE id = ?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		ca.CommonName, ca.Organization, ca.Country, ca.CertificateSHA256,
		ca.TrustedByMozilla, ca.TrustedByMicrosoft, ca.TrustedByChrome,
		ca.EVCapable, time.Now(), ca.ID,
	)
	return err
}

// CAExists checks if a CA already exists by name
func (db *DB) CAExists(name string) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var exists bool
	err := db.conn.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM certificate_authorities WHERE name = ?)",
		name,
	).Scan(&exists)
	return exists, err
}
