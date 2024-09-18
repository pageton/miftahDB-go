package db

// SQLStatements contains all the SQL queries used in the database
var SQLStatements = struct {
	Get          string
	Set          string
	Delete       string
	Cleanup      string
	Rename       string
	CreateTable  string
	CreateIndex  string
	Vacuum       string
	Flush        string
	Exists       string
	GetExpire    string
	SetExpire    string
	Keys         string
	Pagination   string
	CountKeys    string
	CountExpired string
	CreatePragma string
}{
	Get:          "SELECT value, expires_at FROM miftahDB WHERE key = ?",
	Set:          "INSERT OR REPLACE INTO miftahDB (key, value, expires_at) VALUES (?, ?, ?)",
	Delete:       "DELETE FROM miftahDB WHERE key = ?",
	Cleanup:      "DELETE FROM miftahDB WHERE expires_at IS NOT NULL AND expires_at <= ?",
	Rename:       "UPDATE miftahDB SET key = ? WHERE key = ?",
	CreateTable:  `CREATE TABLE IF NOT EXISTS miftahDB (key TEXT PRIMARY KEY, value BLOB, expires_at INTEGER) WITHOUT ROWID;`,
	CreateIndex:  "CREATE INDEX IF NOT EXISTS idx_expires_at ON miftahDB(expires_at) WHERE expires_at IS NOT NULL",
	Vacuum:       "VACUUM",
	Flush:        "DELETE FROM miftahDB",
	Exists:       "SELECT EXISTS (SELECT 1 FROM miftahDB WHERE key = ? LIMIT 1)",
	GetExpire:    "SELECT expires_at FROM miftahDB WHERE key = ?",
	SetExpire:    "UPDATE miftahDB SET expires_at = ? WHERE key = ?",
	Keys:         "SELECT key FROM miftahDB WHERE key LIKE ?",
	Pagination:   "SELECT key FROM miftahDB WHERE key LIKE ? LIMIT ? OFFSET ?",
	CountKeys:    "SELECT COUNT(*) AS count FROM miftahDB where key LIKE ?",
	CountExpired: "SELECT COUNT(*) as count FROM miftahDB WHERE (expires_at IS NOT NULL AND expires_at <= strftime('%s', 'now') * 1000) AND key LIKE ?",
	CreatePragma: `
        PRAGMA journal_mode = WAL;
        PRAGMA synchronous = NORMAL;
        PRAGMA temp_store = MEMORY;
        PRAGMA cache_size = -64000;
        PRAGMA mmap_size = 30000000000;
        PRAGMA optimize;
    `,
}
