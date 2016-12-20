package rdb

import (
	"fmt"
	"reflect"
	"strings"
)

// db maps a database to a map of tables and a
// table that contains a slice of column definitions.
var dbMap = make(map[string]map[string][]column)

// model maps a model struct to a corresponding database/table pair.
var modMap = make(map[string][]string)

// Register scans a model struct and builds a database/table/column map from
// the db struct tags.
//
// All RDB model structs *MUST* define ONE EACH of:
//  - table=tbl_name which maps the name of the table that corresponds with
//    the model struct type being mapped
//  - database=db_name which associates the table with a particular database
//    which is especially necessary to differentiate two tables with the same
//    name as blonging to separate databases.
// And at minimum, at least one struct field must define a column name to pass
// RDB type validation:
//  - col=col_name maps the struct field to a column within the table and database
//    defined.
//
// Database and table names can be defined on any struct field, but must only be
// defined once. Additionally, datbase and table names can be defined separately
// from other columns to improve readability of struct tags, for example:
//
// type RDBModel struct {
// 		_ bool `db:"database=my_database"`
// 		_ bool `db:"table=my_table"`
// 		ID int `db:"col=id,pk,ai"`
// }
//
// Here the field is anonymous and set to type bool to reduce the impact on
// struct size.
//
// Optional flags can be set to define column metaedata:
//  - pk sets a boolean that flags a column as a primary key
//  - ai sets a boolean that flags a column as being an auto-increment column
//    which will let RDB auto-fetch IDs when making insert statements
//  - null sets a boolean that flags whether a column will accept a null value or not
//  - fkmap=ColName.Model.Field maps a struct field that represents an embedded RDB
//    model type defined outside of the model being mapped and tells RDB which
//    column in the table represents the related entity foreign key. fk allows
//    helper functions to load related entities.
func Register(model interface{}) error {
	r := reflect.TypeOf(model)
	if r.Kind() != reflect.Struct {
		return fmt.Errorf("Register can only be called on struct types. Called on %T", model)
	}

	// Read every field in struct for database tags
	// Ignore fields that do not have db tags
	modelName := r.Name()
	nf := r.NumField()
	cols := make([]column, 0, nf)
	colCheck := make(map[string]bool)
	var tblName, dbName string
	for i := 0; i < nf; i++ {
		f := r.Field(i)
		tag, ok := f.Tag.Lookup("db")
		if !ok {
			continue
		}

		col := column{}
		col.colType = f.Type.Kind()
		col.fieldName = f.Name
		colNameSet := false
		dbNameSet := false
		tblNameSet := false

		// Parse database tag for all options
		parts := strings.Split(tag, ",")
		for _, s := range parts {
			s = strings.TrimSpace(s)
			switch {
			// Foreign-key definition
			case len(s) >= 6 && "fkmap=" == s[0:6]:
				cnt := strings.Count(s, ".")
				if cnt != 2 {
					return fmt.Errorf(
						`Foreign-key tag validation error on "%s.%s": Format is "fkmap=ColName.Model.Field" but "%s" given`,
						modelName, f.Name, tag)
				}

				idx := strings.Index(s, ".")
				_, ok := colCheck[s[6:idx]]
				if !ok {
					return fmt.Errorf(
						`Foreign-key column "%s" has not been registered for "%s.%s". "%s" MUST preceed this field in the struct definition.`,
						s[6:idx], modelName, f.Name, s[6:idx])
				}

				col.colRelation = s[6:]
				col.fk = true

			// Table name definition, on PK column by convention
			case len(s) >= 6 && "table=" == s[0:6]:
				if len(s[6:]) == 0 {
					return fmt.Errorf(
						`Table name tag validation error on "%s.%s": Format is "table=table_name"`,
						modelName, f.Name)
				}

				if tblName != "" {
					return fmt.Errorf(
						`Table name cannot be redeclared as "%s" in "%s.%s": already defined as "%s"`,
						s[6:], modelName, f.Name, tblName)
				}

				ind := strings.Index(s, ".")
				if ind >= 0 {
					return fmt.Errorf(
						`Table name validation error on "%s.%s": Name may not include '.', "%s" given`,
						modelName, f.Name, s[6:])
				}

				tblName = s[6:]
				tblNameSet = true

			// Database definition
			case len(s) >= 9 && s[0:9] == "database=":
				if len(s[9:]) == 0 {
					return fmt.Errorf(
						`Database name tag validation error on "%s.%s": Format is "database=db_name"`,
						modelName, f.Name)
				}

				if dbName != "" {
					return fmt.Errorf(
						`Database name cannot be redeclared "%s" in "%s.%s": already defined as "%s"`,
						s[9:], modelName, f.Name, dbName)
				}

				ind := strings.Index(s, ".")
				if ind >= 0 {
					return fmt.Errorf(
						`Database name validation error on "%s.%s": May not include '.' in name, "%s" given`,
						modelName, f.Name, s[9:])
				}

				dbName = s[9:]
				dbNameSet = true

			// Auto-increment definition
			case "ai" == s:
				col.ai = true

			// Primary-key definition
			case "pk" == s:
				col.pk = true

			// Is Nullable definition
			case "null" == s:
				col.null = true

			// Column name definition
			case len(s) >= 4 && s[0:4] == "col=":
				if colNameSet {
					return fmt.Errorf(
						`Cannot use "%s" for column name, "%s" already declared for "%s.%s"`,
						s[4:], col.colName, modelName, f.Name)
				}

				if len(s[4:]) == 0 {
					return fmt.Errorf(
						`Column name validation error on "%s.%s": Format is "col=col_name"`,
						modelName, f.Name)
				}

				if _, ok := colCheck[s[4:]]; ok {
					return fmt.Errorf(
						`Duplicate column name "%s" on "%s.%s", column was already declared on another filed`,
						s[4:], modelName, f.Name)
				}

				col.colName = s[4:]
				colCheck[s[4:]] = true
				colNameSet = true

			// This is an error in every case.
			default:
				return fmt.Errorf(
					`db tag validation error: inspect "%s.%s" for empty db tag values`,
					modelName, f.Name)
			}
		}

		if colNameSet == false && (!dbNameSet && !tblNameSet) {
			return fmt.Errorf(
				`db tag validation error, column name was not found for "%s.%s" in tag: %s`,
				modelName, f.Name, tag)
		}

		if colNameSet {
			cols = append(cols, col)
		}
	}

	if len(cols) == 0 {
		return fmt.Errorf("No columns were defined in %s struct", modelName)
	}

	if dbName == "" {
		return fmt.Errorf("Database name was not defined in %s struct.", modelName)
	}

	if tblName == "" {
		return fmt.Errorf("Table namne was not defined in %s struct.", modelName)
	}

	if _, ok := dbMap[dbName]; ok == false {
		dbMap[dbName] = make(map[string][]column)
	}

	// Everything okay, map model name to db/table for quick Lookup
	// Add table definition to database map
	modMap[modelName] = []string{dbName, tblName}
	dbMap[dbName][tblName] = cols

	return nil
}
