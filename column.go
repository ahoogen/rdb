package rdb

import "reflect"

// column describes the properties related to a column
type column struct {
	fieldName   string       // Model field the column is mapped to
	colName     string       // Table column name in the database
	colRelation string       // Model.Field map for foreign key reference
	colType     reflect.Kind // Type of the column data, not sure this is needed, may be dropped
	pk          bool         // Column is a primary key
	ai          bool         // Column has an auto-incrementer
	fk          bool         // Column is a foreign key
	null        bool         // Column is/is not null
}
