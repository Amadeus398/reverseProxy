package db

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"reverseProxy/pkg/logging"
	"time"
)

var (
	ErrNothingDone = fmt.Errorf("sql query did nothing")
	ConnManager    = AbstractConnectionManager(&ConnectionManager{})
)

type ConnectionManager struct {
	connection *sql.DB
	log        *logging.Logger
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
	c.log = logging.NewLogs("db", "connect")
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
	c.log.GetInfo().Msg("open connection")
	c.connection, err = sql.Open("pgx", connector)
	if err != nil {
		c.log.GetError().Str("when", "open connection").Err(err).Msg("unable to open connection")
		return err
	}
	c.log.GetInfo().Msg("ping connection")
	if err := c.connection.Ping(); err != nil {
		c.log.GetError().Str("when", "ping connection").Err(err).Msg("unable to ping connection")
		return err
	}
	return nil
}

// Close closes a connection to the PostgreSQL server
func (c *ConnectionManager) Close() error {
	c.log = logging.NewLogs("db", "close")
	c.log.GetInfo().Msg("close connection")
	if err := c.connection.Close(); err != nil {
		c.log.GetError().Str("when", "close connection").Err(err).Msg("unable to close connection")
		return err
	}
	return nil
}

// Exec executes a query without returning any rows
func (c *ConnectionManager) Exec(query string, args ...interface{}) error {
	c.log = logging.NewLogs("db", "exec")
	if err := c.connection.Ping(); err != nil {
		c.log.GetError().Str("when", "ping connection").Err(err).Msg("unable to ping connection")
		return err
	}
	ctx := context.TODO()
	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	c.log.GetInfo().Msg("exec")
	result, err := c.connection.ExecContext(queryCtx, query, args...)
	if err != nil {
		c.log.GetError().Str("when", "exec").Err(err).Msg("error at exec")
		return err
	}
	c.log.GetInfo().Msg("get affected rows")
	rows, err := result.RowsAffected()
	if err != nil {
		c.log.GetError().Str("when", "get rows").Err(err).Msg("unable to get rows")
		return err
	}
	if rows == 0 {
		c.log.GetWarn().Str("when", "query did nothing").Err(ErrNothingDone)
		return ErrNothingDone
	}
	return nil
}

// QueryRow executes a query that return at most one row
func (c *ConnectionManager) QueryRow(query string, args ...interface{}) (*sql.Row, func(), error) {
	c.log = logging.NewLogs("db", "queryRow")
	if err := c.connection.Ping(); err != nil {
		c.log.GetError().Str("when", "ping connection").Err(err).Msg("unable to ping connection")
		return nil, nil, err
	}
	ctx := context.TODO()
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

	c.log.GetInfo().Msg("get row")
	row := c.connection.QueryRowContext(queryCtx, query, args...)

	return row, cancel, nil
}

// Query executes a query that returns more than one row
func (c *ConnectionManager) Query(query string, args ...interface{}) (*sql.Rows, func(), error) {
	c.log = logging.NewLogs("db", "query")
	if err := c.connection.Ping(); err != nil {
		c.log.GetError().Str("when", "ping connection").Err(err).Msg("unable to ping connection")
		return nil, nil, err
	}
	ctx := context.TODO()
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

	c.log.GetInfo().Msg("get rows")
	rows, err := c.connection.QueryContext(queryCtx, query, args...)
	if err != nil {
		c.log.GetError().Str("when", "get rows").Err(err).Msg("unable to get rows")
		defer cancel()
		return nil, nil, err
	}
	return rows, cancel, nil
}

type AbstractConnectionManager interface {
	Connect(cfg DbConfig) error
	Close() error
	Exec(query string, args ...interface{}) error
	QueryRow(query string, args ...interface{}) (*sql.Row, func(), error)
	Query(query string, args ...interface{}) (*sql.Rows, func(), error)
}
