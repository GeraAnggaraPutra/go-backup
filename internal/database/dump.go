package database

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"
	"time"
)

func genericDump(ctx context.Context, db *sql.DB, table string, w io.Writer) error {
	stmt := fmt.Sprintf("SELECT * FROM %s", table)
	rows, err := db.QueryContext(ctx, stmt)
	if err != nil {
		return err
	}
	defer rows.Close()

	cols, _ := rows.Columns()

	values := make([]interface{}, len(cols))
	ptrs := make([]interface{}, len(cols))
	for i := range values {
		ptrs[i] = &values[i]
	}

	for rows.Next() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := rows.Scan(ptrs...); err != nil {
			return err
		}

		valStrings := make([]string, len(values))
		for i, val := range values {
			if val == nil {
				valStrings[i] = "NULL"
				continue
			}

			switch v := val.(type) {
			case []byte:
				valStrings[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(string(v), "'", "''"))

			case time.Time:
				valStrings[i] = fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05.000000"))

			case bool:
				if v {
					valStrings[i] = "TRUE"
				} else {
					valStrings[i] = "FALSE"
				}

			case string:
				valStrings[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))

			default:
				valStrings[i] = fmt.Sprintf("%v", v)
			}
		}

		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s);\n",
			table,
			strings.Join(cols, ","),
			strings.Join(valStrings, ","),
		)

		if _, err := w.Write([]byte(query)); err != nil {
			return err
		}
	}

	return nil
}
