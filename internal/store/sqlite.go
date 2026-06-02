package store

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// SQLiteStore Store arayüzünün SQLite implementasyonu.
type SQLiteStore struct {
	db *sql.DB
}

// OpenSQLite verilen DSN ile bağlantı açar ve şemayı uygular.
// DSN örnek: "file:/path/to/mr-ducksafe.db?_journal=WAL&_fk=1"
func OpenSQLite(dsn string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite open: %w", err)
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("schema apply: %w", err)
	}
	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Close() error { return s.db.Close() }

func (s *SQLiteStore) SaveScan(ctx context.Context, sc Scan) (int64, error) {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO scans (started_at, host, os, status)
		VALUES (?, ?, ?, ?)`,
		sc.StartedAt, sc.Host, sc.OS, sc.Status,
	)
	if err != nil {
		return 0, fmt.Errorf("save scan: %w", err)
	}
	return res.LastInsertId()
}

func (s *SQLiteStore) AppendFindings(ctx context.Context, scanID int64, findings []Finding) error {
	if len(findings) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO findings (scan_id, app_name, app_version, cve_id, severity, fixed_version, summary, refs)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, f := range findings {
		refsJSON, err := json.Marshal(f.References)
		if err != nil {
			return fmt.Errorf("refs marshal: %w", err)
		}
		if _, err := stmt.ExecContext(ctx,
			scanID, f.AppName, f.AppVersion, f.CVEID, f.Severity,
			f.FixedVersion, f.Summary, string(refsJSON),
		); err != nil {
			return fmt.Errorf("insert finding: %w", err)
		}
	}
	return tx.Commit()
}

func (s *SQLiteStore) UpdateScanStatus(ctx context.Context, scanID int64, status ScanStatus, finishedAt time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE scans SET status = ?, finished_at = ? WHERE id = ?`,
		status, finishedAt, scanID,
	)
	return err
}

func (s *SQLiteStore) ListScans(ctx context.Context, limit int) ([]Scan, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, started_at, finished_at, host, os, status
		FROM scans ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Scan
	for rows.Next() {
		var sc Scan
		var finished sql.NullTime
		if err := rows.Scan(&sc.ID, &sc.StartedAt, &finished, &sc.Host, &sc.OS, &sc.Status); err != nil {
			return nil, err
		}
		if finished.Valid {
			sc.FinishedAt = &finished.Time
		}
		out = append(out, sc)
	}
	return out, rows.Err()
}

func (s *SQLiteStore) GetScan(ctx context.Context, id int64) (Scan, []Finding, error) {
	var sc Scan
	var finished sql.NullTime
	err := s.db.QueryRowContext(ctx, `
		SELECT id, started_at, finished_at, host, os, status
		FROM scans WHERE id = ?`, id).Scan(
		&sc.ID, &sc.StartedAt, &finished, &sc.Host, &sc.OS, &sc.Status,
	)
	if err != nil {
		return Scan{}, nil, fmt.Errorf("get scan: %w", err)
	}
	if finished.Valid {
		sc.FinishedAt = &finished.Time
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, scan_id, app_name, app_version, cve_id, severity, fixed_version, summary, refs
		FROM findings WHERE scan_id = ? ORDER BY severity, app_name`, id)
	if err != nil {
		return sc, nil, err
	}
	defer rows.Close()

	var findings []Finding
	for rows.Next() {
		var f Finding
		var refsJSON sql.NullString
		var fixed sql.NullString
		if err := rows.Scan(&f.ID, &f.ScanID, &f.AppName, &f.AppVersion, &f.CVEID, &f.Severity, &fixed, &f.Summary, &refsJSON); err != nil {
			return sc, nil, err
		}
		if fixed.Valid {
			f.FixedVersion = fixed.String
		}
		if refsJSON.Valid && refsJSON.String != "" {
			_ = json.Unmarshal([]byte(refsJSON.String), &f.References)
		}
		findings = append(findings, f)
	}
	return sc, findings, rows.Err()
}

func (s *SQLiteStore) PruneOlderThan(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan)
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM scans WHERE started_at < ?`, cutoff)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

func (s *SQLiteStore) CountFindings(ctx context.Context, scanID int64) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM findings WHERE scan_id = ?`, scanID,
	).Scan(&n); err != nil {
		return 0, fmt.Errorf("count findings: %w", err)
	}
	return n, nil
}
