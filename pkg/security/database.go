// Package security provides database security utilities
package security

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// Database security errors
var (
	ErrDatabaseEncryptionRequired = errors.New("database encryption is required in production")
	ErrDatabaseKeyMissing         = errors.New("database encryption key not configured")
	ErrDatabaseCorrupted          = errors.New("database integrity check failed")
	ErrQueryTimeout               = errors.New("database query timed out")
)

// DatabaseSecurityConfig holds database security configuration
type DatabaseSecurityConfig struct {
	// Encryption
	EncryptionEnabled bool
	EncryptionKey     *SecureString

	// Connection limits
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration

	// Query limits
	QueryTimeout    time.Duration
	MaxResultRows   int
	SlowQueryThreshold time.Duration

	// Security settings
	RequireEncryption  bool
	EnableIntegrityCheck bool
	EnableAuditLog     bool
}

// DefaultDatabaseSecurityConfig returns secure defaults
func DefaultDatabaseSecurityConfig() *DatabaseSecurityConfig {
	key := os.Getenv("DB_ENCRYPTION_KEY")
	var encKey *SecureString
	if key != "" {
		encKey = NewSecureString(key)
	}

	return &DatabaseSecurityConfig{
		EncryptionEnabled:    encKey != nil,
		EncryptionKey:        encKey,
		MaxOpenConns:         25,
		MaxIdleConns:         5,
		ConnMaxLifetime:      30 * time.Minute,
		ConnMaxIdleTime:      5 * time.Minute,
		QueryTimeout:         30 * time.Second,
		MaxResultRows:        10000,
		SlowQueryThreshold:   1 * time.Second,
		RequireEncryption:    os.Getenv("REQUIRE_DB_ENCRYPTION") == "true",
		EnableIntegrityCheck: true,
		EnableAuditLog:       true,
	}
}

// SecureDB wraps a database connection with security features
type SecureDB struct {
	db         *sql.DB
	config     *DatabaseSecurityConfig
	queryStats *QueryStats
	auditLog   *DatabaseAuditLog
	mu         sync.RWMutex
}

// QueryStats tracks query statistics for monitoring
type QueryStats struct {
	TotalQueries    int64
	SlowQueries     int64
	FailedQueries   int64
	TotalDuration   time.Duration
	mu              sync.RWMutex
}

// DatabaseAuditLog logs database operations
type DatabaseAuditLog struct {
	entries []AuditEntry
	maxSize int
	mu      sync.RWMutex
}

// AuditEntry represents a database audit log entry
type AuditEntry struct {
	Timestamp time.Time
	Operation string
	Table     string
	UserID    string
	Duration  time.Duration
	Success   bool
	Error     string
}

// NewSecureDB wraps a database connection with security features
func NewSecureDB(db *sql.DB, config *DatabaseSecurityConfig) (*SecureDB, error) {
	if config == nil {
		config = DefaultDatabaseSecurityConfig()
	}

	// Validate configuration
	if config.RequireEncryption && !config.EncryptionEnabled {
		return nil, ErrDatabaseEncryptionRequired
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	sdb := &SecureDB{
		db:         db,
		config:     config,
		queryStats: &QueryStats{},
		auditLog: &DatabaseAuditLog{
			entries: make([]AuditEntry, 0, 1000),
			maxSize: 1000,
		},
	}

	// Run initial integrity check
	if config.EnableIntegrityCheck {
		if err := sdb.IntegrityCheck(context.Background()); err != nil {
			return nil, fmt.Errorf("initial integrity check failed: %w", err)
		}
	}

	return sdb, nil
}

// QueryWithContext executes a query with timeout and logging
func (sdb *SecureDB) QueryWithContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	// Add query timeout
	queryCtx, cancel := context.WithTimeout(ctx, sdb.config.QueryTimeout)
	defer cancel()

	start := time.Now()

	rows, err := sdb.db.QueryContext(queryCtx, query, args...)

	duration := time.Since(start)
	sdb.recordQuery(duration, err)

	// Log slow queries
	if duration > sdb.config.SlowQueryThreshold {
		sdb.logSlowQuery(query, duration)
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return nil, ErrQueryTimeout
	}

	return rows, err
}

// ExecWithContext executes a statement with timeout and logging
func (sdb *SecureDB) ExecWithContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	// Add query timeout
	queryCtx, cancel := context.WithTimeout(ctx, sdb.config.QueryTimeout)
	defer cancel()

	start := time.Now()

	result, err := sdb.db.ExecContext(queryCtx, query, args...)

	duration := time.Since(start)
	sdb.recordQuery(duration, err)

	// Log slow queries
	if duration > sdb.config.SlowQueryThreshold {
		sdb.logSlowQuery(query, duration)
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return nil, ErrQueryTimeout
	}

	return result, err
}

// PrepareStatement creates a prepared statement (SQL injection prevention)
func (sdb *SecureDB) PrepareStatement(ctx context.Context, query string) (*sql.Stmt, error) {
	// Validate query doesn't contain obvious SQL injection patterns
	if sdb.detectSQLInjection(query) {
		return nil, errors.New("potential SQL injection detected in query")
	}

	return sdb.db.PrepareContext(ctx, query)
}

func (sdb *SecureDB) detectSQLInjection(query string) bool {
	// Check for common SQL injection patterns in the query template itself
	// This is a safety check, not the primary defense (prepared statements are)
	suspicious := []string{
		"--",           // SQL comment
		"/*",           // Block comment start
		"*/",           // Block comment end
		";--",          // Statement terminator with comment
		"'--",          // String terminator with comment
		"\"--",         // Double quote with comment
		" OR 1=1",      // Classic injection
		" OR '1'='1",   // Variant
		"UNION SELECT", // Union injection
		"DROP TABLE",   // Destructive
		"DELETE FROM",  // Destructive without WHERE
		"TRUNCATE",     // Destructive
		"xp_",          // SQL Server extended procedures
	}

	upperQuery := strings.ToUpper(query)
	for _, pattern := range suspicious {
		if strings.Contains(upperQuery, strings.ToUpper(pattern)) {
			return true
		}
	}

	return false
}

// IntegrityCheck performs a database integrity check
func (sdb *SecureDB) IntegrityCheck(ctx context.Context) error {
	// SQLite specific integrity check
	rows, err := sdb.db.QueryContext(ctx, "PRAGMA integrity_check")
	if err != nil {
		return err
	}
	defer rows.Close()

	var result string
	if rows.Next() {
		if err := rows.Scan(&result); err != nil {
			return err
		}
	}

	if result != "ok" {
		return fmt.Errorf("%w: %s", ErrDatabaseCorrupted, result)
	}

	return nil
}

// BeginSecureTx starts a secure transaction with proper isolation
func (sdb *SecureDB) BeginSecureTx(ctx context.Context, opts *sql.TxOptions) (*SecureTx, error) {
	if opts == nil {
		opts = &sql.TxOptions{
			Isolation: sql.LevelSerializable,
			ReadOnly:  false,
		}
	}

	tx, err := sdb.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &SecureTx{
		tx:     tx,
		config: sdb.config,
		start:  time.Now(),
	}, nil
}

func (sdb *SecureDB) recordQuery(duration time.Duration, err error) {
	sdb.queryStats.mu.Lock()
	defer sdb.queryStats.mu.Unlock()

	sdb.queryStats.TotalQueries++
	sdb.queryStats.TotalDuration += duration

	if err != nil {
		sdb.queryStats.FailedQueries++
	}

	if duration > sdb.config.SlowQueryThreshold {
		sdb.queryStats.SlowQueries++
	}
}

func (sdb *SecureDB) logSlowQuery(query string, duration time.Duration) {
	// Truncate query for logging
	maxLen := 200
	if len(query) > maxLen {
		query = query[:maxLen] + "..."
	}

	// In production, use structured logging
	// log.Printf("[DB SLOW QUERY] %s took %s", query, duration)
}

// GetStats returns query statistics
func (sdb *SecureDB) GetStats() QueryStats {
	sdb.queryStats.mu.RLock()
	defer sdb.queryStats.mu.RUnlock()

	return QueryStats{
		TotalQueries:  sdb.queryStats.TotalQueries,
		SlowQueries:   sdb.queryStats.SlowQueries,
		FailedQueries: sdb.queryStats.FailedQueries,
		TotalDuration: sdb.queryStats.TotalDuration,
	}
}

// AuditLog logs a database operation
func (sdb *SecureDB) AuditLog(entry AuditEntry) {
	if !sdb.config.EnableAuditLog {
		return
	}

	sdb.auditLog.mu.Lock()
	defer sdb.auditLog.mu.Unlock()

	if len(sdb.auditLog.entries) >= sdb.auditLog.maxSize {
		// Remove oldest entries
		sdb.auditLog.entries = sdb.auditLog.entries[100:]
	}

	sdb.auditLog.entries = append(sdb.auditLog.entries, entry)
}

// GetAuditLog returns recent audit log entries
func (sdb *SecureDB) GetAuditLog(limit int) []AuditEntry {
	sdb.auditLog.mu.RLock()
	defer sdb.auditLog.mu.RUnlock()

	if limit <= 0 || limit > len(sdb.auditLog.entries) {
		limit = len(sdb.auditLog.entries)
	}

	start := len(sdb.auditLog.entries) - limit
	result := make([]AuditEntry, limit)
	copy(result, sdb.auditLog.entries[start:])
	return result
}

// Close closes the database connection
func (sdb *SecureDB) Close() error {
	// Wipe encryption key if present
	if sdb.config.EncryptionKey != nil {
		sdb.config.EncryptionKey.Wipe()
	}

	return sdb.db.Close()
}

// SecureTx represents a secure database transaction
type SecureTx struct {
	tx     *sql.Tx
	config *DatabaseSecurityConfig
	start  time.Time
}

// Commit commits the transaction
func (stx *SecureTx) Commit() error {
	return stx.tx.Commit()
}

// Rollback rolls back the transaction
func (stx *SecureTx) Rollback() error {
	return stx.tx.Rollback()
}

// ExecContext executes a statement within the transaction
func (stx *SecureTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return stx.tx.ExecContext(ctx, query, args...)
}

// QueryContext executes a query within the transaction
func (stx *SecureTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return stx.tx.QueryContext(ctx, query, args...)
}

// DataRetentionManager handles automatic data purging
type DataRetentionManager struct {
	db       *SecureDB
	policies []RetentionPolicy
	mu       sync.RWMutex
}

// RetentionPolicy defines a data retention policy
type RetentionPolicy struct {
	Name        string
	Table       string
	DateColumn  string
	MaxAge      time.Duration
	DeleteQuery string
}

// NewDataRetentionManager creates a new data retention manager
func NewDataRetentionManager(db *SecureDB) *DataRetentionManager {
	return &DataRetentionManager{
		db:       db,
		policies: make([]RetentionPolicy, 0),
	}
}

// AddPolicy adds a retention policy
func (drm *DataRetentionManager) AddPolicy(policy RetentionPolicy) {
	drm.mu.Lock()
	defer drm.mu.Unlock()
	drm.policies = append(drm.policies, policy)
}

// EnforceAll enforces all retention policies
func (drm *DataRetentionManager) EnforceAll(ctx context.Context) error {
	drm.mu.RLock()
	policies := make([]RetentionPolicy, len(drm.policies))
	copy(policies, drm.policies)
	drm.mu.RUnlock()

	for _, policy := range policies {
		if err := drm.enforce(ctx, policy); err != nil {
			return fmt.Errorf("failed to enforce policy %s: %w", policy.Name, err)
		}
	}

	return nil
}

func (drm *DataRetentionManager) enforce(ctx context.Context, policy RetentionPolicy) error {
	query := policy.DeleteQuery
	if query == "" {
		cutoff := time.Now().Add(-policy.MaxAge)
		query = fmt.Sprintf("DELETE FROM %s WHERE %s < ?", policy.Table, policy.DateColumn)
		_, err := drm.db.ExecWithContext(ctx, query, cutoff)
		return err
	}

	_, err := drm.db.ExecWithContext(ctx, query)
	return err
}
