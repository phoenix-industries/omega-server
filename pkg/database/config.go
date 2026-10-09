package database

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config describes Postgres connection config.
type Config struct {
	Name               string
	User               string
	Host               string
	Port               string
	Password           string
	ConnectionTimeout  int
	SSLMode            string
	SSLKeyPath         string
	SSLCertPath        string
	SSLRootCertPath    string
	PoolMinConnections string
	PoolMaxConnections string
	PoolMaxConnLife    time.Duration
	PoolMaxConnIdle    time.Duration
	PoolHealthCheck    time.Duration
}

// ConfigFromEnv returns database config based on environment variables.
//
// returns error if POSTGRES_CONNECT_TIMEOUT is not a valid int
// or POSTGRES_POOL_MAX_CONN_LIFETIME, DB_POOL_MAX_CONN_IDLE_TIME
// are not valid time.Duration.
func ConfigFromEnv() (*Config, error) {
	timeout, err := strconv.Atoi(os.Getenv("POSTGRES_CONNECT_TIMEOUT"))
	if err != nil {
		return nil, err
	}

	maxLife, err := time.ParseDuration(os.Getenv("POSTGRES_POOL_MAX_CONN_LIFETIME"))
	if err != nil {
		return nil, err
	}

	maxIdle, err := time.ParseDuration(os.Getenv("POSTGRES_POOL_MAX_CONN_IDLE_TIME"))
	if err != nil {
		return nil, err
	}

	healthCheck, err := time.ParseDuration(os.Getenv("POSTGRES_POOL_HEALTH_CHECK_PERIOD"))
	if err != nil {
		return nil, err
	}

	return &Config{
		Name:               os.Getenv("POSTGRES_DB"),
		User:               os.Getenv("POSTGRES_USER"),
		Host:               os.Getenv("POSTGRES_HOST"),
		Port:               os.Getenv("POSTGRES_PORT"),
		Password:           os.Getenv("POSTGRES_PASSWORD"),
		ConnectionTimeout:  timeout,
		SSLMode:            os.Getenv("POSTGRES_SSLMODE"),
		SSLKeyPath:         os.Getenv("POSTGRES_SSLKEY"),
		SSLCertPath:        os.Getenv("POSTGRES_SSLCERT"),
		SSLRootCertPath:    os.Getenv("POSTGRES_SSLROOTCERT"),
		PoolMinConnections: os.Getenv("POSTGRES_POOL_MIN_CONNS"),
		PoolMaxConnections: os.Getenv("POSTGRES_POOL_MAX_CONNS"),
		PoolMaxConnLife:    maxLife,
		PoolMaxConnIdle:    maxIdle,
		PoolHealthCheck:    healthCheck,
	}, nil
}

func (c *Config) ConnectionString() string {
	vals := c.dbValues()
	p := make([]string, 0, len(vals))
	for k, v := range vals {
		if k == "password" {
			p = append(p, fmt.Sprintf("%s=%s", k, strings.Repeat("*", len(v))))
		}
		p = append(p, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(p, " ")
}

func (c *Config) dbValues() map[string]string {
	m := map[string]string{}
	setIfNotEmpty(m, "dbname", c.Name)
	setIfNotEmpty(m, "user", c.User)
	setIfNotEmpty(m, "host", c.Host)
	setIfNotEmpty(m, "port", c.Port)
	setIfNotEmpty(m, "sslmode", c.SSLMode)
	setIfPositive(m, "connect_timeout", c.ConnectionTimeout)
	setIfNotEmpty(m, "password", c.Password)
	setIfNotEmpty(m, "sslcert", c.SSLCertPath)
	setIfNotEmpty(m, "sslkey", c.SSLKeyPath)
	setIfNotEmpty(m, "sslrootcert", c.SSLRootCertPath)
	setIfNotEmpty(m, "pool_min_conns", c.PoolMinConnections)
	setIfNotEmpty(m, "pool_max_conns", c.PoolMaxConnections)
	setIfPositiveDuration(m, "pool_max_conn_lifetime", c.PoolMaxConnLife)
	setIfPositiveDuration(m, "pool_max_conn_idle_time", c.PoolMaxConnIdle)
	setIfPositiveDuration(m, "pool_health_check_period", c.PoolHealthCheck)
	return m
}

func setIfNotEmpty(m map[string]string, k, v string) {
	if v != "" {
		m[k] = v
	}
}

func setIfPositive(m map[string]string, k string, v int) {
	if v > 0 {
		m[k] = strconv.Itoa(v)
	}
}

func setIfPositiveDuration(m map[string]string, k string, v time.Duration) {
	if v > 0 {
		m[k] = v.String()
	}
}
