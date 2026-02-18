package database

import (
	"context"
	"database/sql"
	"fmt"
	"io"

	_ "github.com/lib/pq"
)

type Postgres struct{}

func (p *Postgres) Connect(host string, port int, user, pass, dbname string) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		user, pass, host, port, dbname,
	)

	return sql.Open("postgres", dsn)
}

func (p *Postgres) ListTables(ctx context.Context, db *sql.DB) ([]string, error) {
	const stmt = `
		SELECT 
			tablename
		FROM 
			pg_tables
		WHERE 
			schemaname='public'
	`

	rows, err := db.QueryContext(ctx, stmt)
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

func (p *Postgres) DumpTable(ctx context.Context, db *sql.DB, table string, w io.Writer) error {
	return genericDump(ctx, db, table, w)
}
