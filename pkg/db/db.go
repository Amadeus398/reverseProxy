package db

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"time"
)

var (
	ErrNothingDone = fmt.Errorf("sql query did nothing")
	ConnManager    = AbstractConnectionManager(&ConnectionManager{})
)

type ConnectionManager struct {
	connection *sql.DB
}

type sqlInfo struct {
	host     string
	port     string
	user     string
	password string
	dbname   string
	sslmode  string
}

type DbConfig interface {
	GetHost() string
	GetPort() string
	GetUser() string
	GetPassword() string
	GetDbname() string
	GetSslmode() string
}

// Connect opens a connection to the PostgreSQL server
func (c *ConnectionManager) Connect(cfg DbConfig) error {
	connSql := sqlInfo{
		host:     cfg.GetHost(),
		port:     cfg.GetPort(),
		user:     cfg.GetUser(),
		password: cfg.GetPassword(),
		dbname:   cfg.GetDbname(),
		sslmode:  cfg.GetSslmode(),
	}
	connector := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		connSql.host, connSql.port, connSql.user, connSql.password, connSql.dbname, connSql.sslmode,
	)
	var err error
	c.connection, err = sql.Open("pgx", connector)
	if err != nil {
		return err
	}
	if err := c.connection.Ping(); err != nil {
		return err
	}
	return nil
}

// Close closes a connection to the PostgreSQL server
func (c *ConnectionManager) Close() {
	if err := c.connection.Close(); err != nil {
		panic(err)
	}
}

// Exec executes a query without returning any rows
func (c *ConnectionManager) Exec(query string, args ...interface{}) error {
	if err := c.connection.Ping(); err != nil {
		return err
	}
	ctx := context.TODO()
	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	result, err := c.connection.ExecContext(queryCtx, query, args...)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNothingDone
	}
	return nil
}

// QueryRow executes a query that return at most one row
func (c *ConnectionManager) QueryRow(query string, args ...interface{}) (*sql.Row, func(), error) {
	if err := c.connection.Ping(); err != nil {
		return nil, nil, err
	}
	ctx := context.TODO()
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

	row := c.connection.QueryRowContext(queryCtx, query, args...)
	return row, cancel, nil

}

// Query executes a query that returns more than one row
func (c *ConnectionManager) Query(query string, args ...interface{}) (*sql.Rows, func(), error) {
	if err := c.connection.Ping(); err != nil {
		return nil, nil, err
	}
	ctx := context.TODO()
	queryCtx, cancel := context.WithTimeout(ctx, 50*time.Second)

	rows, err := c.connection.QueryContext(queryCtx, query, args...)
	if err != nil {
		defer cancel()
		return nil, nil, err
	}
	return rows, cancel, nil
}

type AbstractConnectionManager interface {
	Connect(cfg DbConfig) error
	Close()
	Exec(query string, args ...interface{}) error
	QueryRow(query string, args ...interface{}) (*sql.Row, func(), error)
	Query(query string, args ...interface{}) (*sql.Rows, func(), error)
}
