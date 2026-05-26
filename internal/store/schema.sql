CREATE TABLE IF NOT EXISTS scans (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    started_at   DATETIME NOT NULL,
    finished_at  DATETIME,
    host         TEXT,
    os           TEXT,
    status       TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS findings (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    scan_id       INTEGER NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    app_name      TEXT NOT NULL,
    app_version   TEXT NOT NULL,
    cve_id        TEXT NOT NULL,
    severity      TEXT NOT NULL,
    fixed_version TEXT,
    summary       TEXT,
    refs          TEXT
);

CREATE INDEX IF NOT EXISTS idx_findings_scan_id ON findings(scan_id);
CREATE INDEX IF NOT EXISTS idx_findings_severity ON findings(severity);
