package db

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
	"time"
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

var (
	ErrNothingDone = fmt.Errorf("sql query did nothing")
	ConnManager    = ConnectionManager{}
	connSql        = sqlInfo{
		host:     "127.0.0.1",
		port:     "5432",
		user:     "amadeus",
		password: "digger",
		dbname:   "log",
		sslmode:  "disable",
	}
)

func (c *ConnectionManager) Connect() error {
	connector := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		connSql.host, connSql.port, connSql.user, connSql.password, connSql.dbname, connSql.sslmode,
	)
	var err error
	c.connection, err = sql.Open("pgx", connector)
	if err != nil {
		//TODO zerolog
		log.Fatal(err)
	}
	if err := c.connection.Ping(); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (c *ConnectionManager) Close() {
	if err := c.connection.Close(); err != nil {
		// TODO zerolog
		return
	}
}

func (c *ConnectionManager) Exec(query string, args ...interface{}) error {
	if err := c.connection.Ping(); err != nil {
		return err
	}
	ctx := context.TODO()
	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	result, err := c.connection.ExecContext(queryCtx, query, args...)
	if err != nil {
		// TODO zerolog
		fmt.Println(err)
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

func (c *ConnectionManager) QueryRow(query string, args ...interface{}) (*sql.Row, func(), error) {
	if err := c.connection.Ping(); err != nil {
		return nil, nil, err
	}
	ctx := context.TODO()
	queryCtx, cancel := context.WithTimeout(ctx, 50*time.Second)

	row := c.connection.QueryRowContext(queryCtx, query, args...)
	return row, cancel, nil

}

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
