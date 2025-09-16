package storage

import (
	"context"
	"cyberark-shorten-url/internal/model"
	"cyberark-shorten-url/pkg/logger"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	DefaultMaxOpenConns    = 50
	DefaultMaxIdleConns    = 25
	DefaultConnMaxLifetime = 5 * time.Minute

	DefaultPingTimeout    = 3 * time.Second
	DefaultRequestTimeout = 3 * time.Second
)

type DbClient struct {
	DB      *sql.DB
	BaseCtx context.Context
}

// NewDbClient opens a MySQL connection, configures the pool, pings the DB with a short timeout,
// and returns a pointer to DbClient.
func NewDbClient(dbconfig *model.DatabaseConfig) *DbClient {
	connStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		dbconfig.Username,
		dbconfig.Password,
		dbconfig.DatabaseHost,
		dbconfig.DatabaseName,
	)
	logger.Info(fmt.Sprintf("database connection string: %s", connStr))

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to connect to database: %s", err.Error()))
		return nil
	}

	// Connection pool configuration. Use default value for interview test.
	db.SetMaxOpenConns(DefaultMaxOpenConns)
	db.SetMaxIdleConns(DefaultMaxIdleConns)
	db.SetConnMaxLifetime(DefaultConnMaxLifetime)

	// Verify connectivity with a short timeout.
	pingCtx, cancel := context.WithTimeout(context.Background(), DefaultPingTimeout)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		logger.Fatal(fmt.Sprintf("verify database connect failed: %s", err.Error()))
		return nil
	}

	logger.Info("db client initialize complete")
	return &DbClient{
		DB:      db,
		BaseCtx: context.Background(),
	}
}

// NewRequestContext returns a request-scoped context with timeout.
// Always defer the cancel function in the caller.
//
// If timeout <= 0, DefaultRequestTimeout is used.
func (m *DbClient) NewRequestContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	if m == nil || m.BaseCtx == nil {
		// Fallback to Background if model is nil or BaseCtx not set.
		if timeout <= 0 {
			timeout = DefaultRequestTimeout
		}
		return context.WithTimeout(context.Background(), timeout)
	}
	if timeout <= 0 {
		timeout = DefaultRequestTimeout
	}
	return context.WithTimeout(m.BaseCtx, timeout)
}

// Close closes the underlying DB connection pool.
func (m *DbClient) Close() {
	if m == nil || m.DB == nil {
		return
	}
	m.DB.Close()

	logger.Info("db client shutdown complete")
}
