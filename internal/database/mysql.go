package database

import (
	"context"
	"database/sql"
	"fmt"
	"io"

	_ "github.com/go-sql-driver/mysql"
)

type MySQL struct{}

func (m *MySQL) Connect(host string, port int, user, pass, dbname string) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		user, pass, host, port, dbname,
	)

	return sql.Open("mysql", dsn)
}

func (m *MySQL) ListTables(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	var name string

	for rows.Next() {
		rows.Scan(&name)
		tables = append(tables, name)
	}

	return tables, nil
}

func (m *MySQL) DumpTable(ctx context.Context, db *sql.DB, table string, w io.Writer) error {
	return genericDump(ctx, db, table, w)
}
