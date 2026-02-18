package backup

import "time"

type TableProgress struct {
	Table     string
	Bytes     int64
	Done      bool
	StartTime time.Time
}
