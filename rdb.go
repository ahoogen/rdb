package rdb

import "database/sql"

// Rdb is the package access point for relational database operations.
type Rdb struct {
	Db *sql.DB
}
