package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Connection pool metrics
var (
	poolOpenConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_pool_open_connections",
		Help: "Number of open database connections",
	})

	poolInUseConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_pool_in_use_connections",
		Help: "Number of in-use database connections",
	})

	poolIdleConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_pool_idle_connections",
		Help: "Number of idle database connections",
	})

	poolWaitDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "db_pool_wait_duration_seconds",
		Help:    "Wait duration for a database connection",
		Buckets: prometheus.DefBuckets,
	})

	poolMaxOpenConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_pool_max_open_connections",
		Help: "Maximum number of open database connections",
	})

	healthCheckCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_health_check_total",
			Help: "Total number of database health checks",
		},
		[]string{"status"},
	)
)

// PoolConfig holds database connection pool configuration
type PoolConfig struct {
	MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS" default:"10"`
	ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" default:"30m"`
	ConnMaxIdleTime time.Duration `env:"DB_CONN_MAX_IDLE_TIME" default:"10m"`
	HealthCheckInterval time.Duration `env:"DB_HEALTH_CHECK_INTERVAL" default:"5m"`
}

// DefaultPoolConfig returns default pool configuration
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxOpenConns:        25,
		MaxIdleConns:        10,
		ConnMaxLifetime:     30 * time.Minute,
		ConnMaxIdleTime:     10 * time.Minute,
		HealthCheckInterval: 5 * time.Minute,
	}
}

// Pool wraps sql.DB with monitoring and health checking
type Pool struct {
	*sql.DB
	config PoolConfig
	quit   chan struct{}
}

// NewPool creates a new database connection pool
func NewPool(connStr string, config PoolConfig) (*Pool, error) {
	// Parse connection string for SSL mode
	if os.Getenv("DB_SSL_MODE") == "" {
		connStr += " sslmode=require"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	pool := &Pool{
		DB:     db,
		config: config,
		quit:   make(chan struct{}),
	}

	// Start metrics collector
	go pool.collectMetrics()

	// Start health checker
	go pool.healthChecker()

	return pool, nil
}

// collectMetrics periodically collects pool statistics
func (p *Pool) collectMetrics() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := p.Stats()
			poolOpenConnections.Set(float64(stats.OpenConnections))
			poolInUseConnections.Set(float64(stats.InUse))
			poolIdleConnections.Set(float64(stats.Idle))
			poolMaxOpenConnections.Set(float64(stats.MaxOpenConnections))
			poolWaitDuration.Observe(stats.WaitDuration.Seconds())
		case <-p.quit:
			return
		}
	}
}

// healthChecker performs periodic health checks
func (p *Pool) healthChecker() {
	ticker := time.NewTicker(p.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := p.PingContext(ctx)
			cancel()

			if err != nil {
				log.Printf("[WARN] Database health check failed: %v", err)
				healthCheckCounter.WithLabelValues("fail").Inc()
			} else {
				healthCheckCounter.WithLabelValues("ok").Inc()
			}
		case <-p.quit:
			return
		}
	}
}

// Close gracefully shuts down the pool
func (p *Pool) Close() error {
	close(p.quit)
	return p.DB.Close()
}

// SetUserContext sets the current user ID and tenant ID for RLS
func (p *Pool) SetUserContext(ctx context.Context, userID, tenantID string) (*sql.Conn, error) {
	conn, err := p.Conn(ctx)
	if err != nil {
		return nil, err
	}

	// Set session variables for RLS
	if userID != "" {
		_, err = conn.ExecContext(ctx, fmt.Sprintf("SET app.current_user_id = '%s'", userID))
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to set user_id: %w", err)
		}
	}

	if tenantID != "" {
		_, err = conn.ExecContext(ctx, fmt.Sprintf("SET app.current_tenant_id = '%s'", tenantID))
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to set tenant_id: %w", err)
		}
	}

	return conn, nil
}

// WithTransaction executes a function within a transaction
func (p *Pool) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := p.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// IsHealthy checks if the database connection is healthy
func (p *Pool) IsHealthy(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var one int
	err := p.QueryRowContext(ctx, "SELECT 1").Scan(&one)
	return err == nil
}
