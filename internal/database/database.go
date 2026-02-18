package database

import (
	"context"
	"database/sql"
	"io"
)

type Database interface {
	Connect(host string, port int, user, pass, dbname string) (*sql.DB, error)
	ListTables(ctx context.Context, db *sql.DB) ([]string, error)
	DumpTable(ctx context.Context, db *sql.DB, table string, w io.Writer) error
}
