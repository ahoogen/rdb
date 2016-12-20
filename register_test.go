package rdb

import (
	"fmt"
	"reflect"
	"testing"
)

func reset() {
	dbMap = make(map[string]map[string][]column)
	modMap = make(map[string][]string)
}

func TestNotOnStructError(t *testing.T) {
	defer reset()
	var b bool
	var e error
	if e = Register(b); e == nil {
		t.Errorf("Expected error on registering non-struct value")
	} else {
		m := "Register can only be called on struct types. Called on bool"
		if e.Error() != m {
			t.Errorf("Expected error message '%s'\nGot error message '%s' instead", m, e.Error())
		}
	}
}

func TestOnEmptyStructError(t *testing.T) {
	defer reset()
	type empty struct{}
	var e error
	if e = Register(empty{}); e == nil {
		t.Errorf("Expecting error on TestOnEmptyStructError")
	} else {
		m := "No columns were defined in empty struct"
		if e.Error() != m {
			t.Errorf("Expected error message '%s'\nGot error message '%s' instead", m, e.Error())
		}
	}
}

func TestMissingDatabaseName(t *testing.T) {
	defer reset()
	type missingDB struct {
		ID int `db:"col=test"`
	}
	var e error
	if e = Register(missingDB{}); e == nil {
		t.Errorf("Expecting error on TestMissingDatabaseName")
	} else {
		m := "Database name was not defined in missingDB struct."
		if e.Error() != m {
			t.Errorf("Expected error message '%s'\nGot error message '%s' instead", m, e.Error())
		}
	}
}

func TestMissingTableName(t *testing.T) {
	defer reset()
	type missingTbl struct {
		ID int `db:"col=test,database=test"`
	}
	var e error
	if e = Register(missingTbl{}); e == nil {
		t.Errorf("Expecting error on TestMissingTableName")
	} else {
		m := "Table namne was not defined in missingTbl struct."
		if e.Error() != m {
			t.Errorf("Expected error message '%s'\nGot error message '%s' instead", m, e.Error())
		}
	}
}

func TestEmptyDBTags(t *testing.T) {
	defer reset()
	type emptyDBTag struct {
		ID int `db:""`
	}
	var e error
	if e = Register(emptyDBTag{}); e == nil {
		t.Fail()
	} else {
		m := `db tag validation error: inspect "emptyDBTag.ID" for empty db tag values`
		if e.Error() != m {
			t.Errorf("Expected:\n'%s'\nGot:\n'%s'", m, e.Error())
		}
	}
}

func TestMissingColName(t *testing.T) {
	defer reset()
	type missingCol struct {
		ID int `db:"pk,ai"`
	}
	var e error
	if e = Register(missingCol{}); e == nil {
		t.Errorf("Expecting error on TestMissingColName")
	} else {
		m := `db tag validation error, column name was not found for "missingCol.ID" in tag: pk,ai`
		if e.Error() != m {
			t.Errorf("Expected:\n'%s'\nGot:\n'%s'", m, e.Error())
		}
	}
}

func TestDuplicateColNameError(t *testing.T) {
	defer reset()
	type duplicateCol struct {
		ID  int `db:"database=foo,table=bar,col=baz"`
		DUP int `db:"col=baz"`
	}
	var e error
	if e = Register(duplicateCol{}); e == nil {
		t.Errorf("Expecting error on TestDuplicateColNameError")
	} else {
		m := `Duplicate column name "baz" on "duplicateCol.DUP", column was already declared on another filed`
		if e.Error() != m {
			t.Errorf("Expected:\n'%s'\nGot:\n'%s'", m, e.Error())
		}
	}
}

func TestColumnFormatError(t *testing.T) {
	defer reset()
	type badColumnFormat struct {
		ID int `db:"database=foo,table=bar,col="`
	}
	var e error
	if e = Register(badColumnFormat{}); e == nil {
		t.Errorf("Expecting error on TestColumnFormatError")
	} else {
		m := `Column name validation error on "badColumnFormat.ID": Format is "col=col_name"`
		if e.Error() != m {
			t.Errorf("Expected:\n'%s'\nGot:\n'%s'", m, e.Error())
		}
	}
}

func TestColumnNameAlreadySet(t *testing.T) {
	defer reset()
	type colNameAlreadySet struct {
		ID int `db:"col=foo,database=bar,table=baz,col=bang"`
	}
	var e error
	if e = Register(colNameAlreadySet{}); e == nil {
		t.Errorf("Expecting error on TestColumnNameAlreadySet")
	} else {
		m := `Cannot use "bang" for column name, "foo" already declared for "colNameAlreadySet.ID"`
		if e.Error() != m {
			t.Errorf("Expected:\n'%s'\nGot:\n'%s'", m, e.Error())
		}
	}
}

func TestPeriodNotAllowedInDatabaseName(t *testing.T) {
	defer reset()
	type databaseNameWithPeriod struct {
		ID int `db:"col=foo,database=bar.baz,table=bang"`
	}
	var e error
	if e = Register(databaseNameWithPeriod{}); e == nil {
		t.Errorf("Expecting error on TestPeriodNotAllowedInDatabaseName")
	} else {
		m := `Database name validation error on "databaseNameWithPeriod.ID": May not include '.' in name, "bar.baz" given`
		if e.Error() != m {
			t.Errorf("Expected:\n'%s'\nGot:\n'%s'", m, e.Error())
		}
	}
}

func TestDatabaseReDeclaredError(t *testing.T) {
	defer reset()
	type databaseRedeclared struct {
		ID int `db:"database=foo,table=bar,database=baz"`
	}
	var e error
	if e = Register(databaseRedeclared{}); e == nil {
		t.Errorf("Expecting error on TestDatabaseReDeclaredError")
	} else {
		m := `Database name cannot be redeclared "baz" in "databaseRedeclared.ID": already defined as "foo"`
		if e.Error() != m {
			t.Errorf("Expected:\n'%s'\nGot:\n'%s'", m, e.Error())
		}
	}
}

func TestDatabaseNameValidationError(t *testing.T) {
	defer reset()
	type databaseNameValidation struct {
		ID int `db:"database=,table=bar,col=baz"`
	}
	var e error
	if e = Register(databaseNameValidation{}); e == nil {
		t.Errorf("Expecting error on TestDatabaseNameValidationError")
	} else {
		m := `Database name tag validation error on "databaseNameValidation.ID": Format is "database=db_name"`
		if e.Error() != m {
			t.Errorf("Expected:\n'%s'\nGot:\n'%s'", m, e.Error())
		}
	}
}

func TestPeriodNotAllowedInTableName(t *testing.T) {
	defer reset()
	type tableNameWithPeriod struct {
		ID int `db:"col=foo,database=bar,table=baz.bang"`
	}
	var e error
	if e = Register(tableNameWithPeriod{}); e == nil {
		t.Errorf("Expecting error on TestPeriodNotAllowedInTableName")
	} else {
		m := `Table name validation error on "tableNameWithPeriod.ID": Name may not include '.', "baz.bang" given`
		if e.Error() != m {
			t.Errorf("Expected:\n'%s'\nGot:\n'%s'", m, e.Error())
		}
	}
}

func TestTableReDeclaredError(t *testing.T) {
	defer reset()
	type tableRedeclared struct {
		ID int `db:"table=foo,table=bar,database=baz"`
	}
	var e error
	if e = Register(tableRedeclared{}); e == nil {
		t.Errorf("Expecting error on TestDatabaseReDeclaredError")
	} else {
		m := `Table name cannot be redeclared as "bar" in "tableRedeclared.ID": already defined as "foo"`
		if e.Error() != m {
			t.Errorf("Expected:\n'%s'\nGot:\n'%s'", m, e.Error())
		}
	}
}

func TestTableNameValidationError(t *testing.T) {
	defer reset()
	type tableNameValidation struct {
		ID int `db:"database=foo,table=,col=baz"`
	}
	var e error
	if e = Register(tableNameValidation{}); e == nil {
		t.Errorf("Expecting error on TestDatabaseNameValidationError")
	} else {
		m := `Table name tag validation error on "tableNameValidation.ID": Format is "table=table_name"`
		if e.Error() != m {
			t.Errorf("Expected:\n'%s'\nGot:\n'%s'", m, e.Error())
		}
	}
}

func TestForeignKeyMissingSeparator(t *testing.T) {
	defer reset()
	type foreignKeyMissingSeparator struct {
		ID int `db:"fkmap=MissingFieldSeparators"`
	}
	var e error
	if e = Register(foreignKeyMissingSeparator{}); e == nil {
		t.Errorf("Expecting error on TestForeignKeyMissingSeparator")
	} else {
		m := `Foreign-key tag validation error on "foreignKeyMissingSeparator.ID": Format is "fkmap=ColName.Model.Field" but "fkmap=MissingFieldSeparators" given`
		if e.Error() != m {
			t.Errorf("Expected:\n'%s'\nGot:\n'%s'", m, e.Error())
		}
	}
}

func TestForeignKeyColumnNotFound(t *testing.T) {
	defer reset()
	type foreignKeyColumnNotFound struct {
		ID int `db:"fkmap=Poots.Pots.Peets"`
	}
	var e error
	if e = Register(foreignKeyColumnNotFound{}); e == nil {
		t.Errorf("Expecting error on TestForeignKeyColumnNotFound")
	} else {
		m := `Foreign-key column "Poots" has not been registered for "foreignKeyColumnNotFound.ID". "Poots" MUST preceed this field in the struct definition.`
		if e.Error() != m {
			t.Errorf("Expected:\n'%s'\nGot:\n'%s'", m, e.Error())
		}
	}
}

func TestForeignKeyValidationError(t *testing.T) {
	defer reset()
	type foriegnKeyValidationError struct {
		ID int `db:"fkmap=MissingColName.MissingFieldSeparators"`
	}
	var e error
	if e = Register(foriegnKeyValidationError{}); e == nil {
		t.Errorf("Expecting error on TestForeignKeyValidationError")
	} else {
		m := `Foreign-key tag validation error on "foriegnKeyValidationError.ID": Format is "fkmap=ColName.Model.Field" but "fkmap=MissingColName.MissingFieldSeparators" given`
		if e.Error() != m {
			t.Errorf("Expected:\n'%s'\nGot:\n'%s'", m, e.Error())
		}
	}
}

func TestCoverageOnNoDbFieldSkip(t *testing.T) {
	defer reset()
	type foriegnKeyValidationError struct {
		non bool `non-db:"I'm not a DB tag..."`
		ID  int  `db:"database=foo,table=bar,col=baz"`
	}
	if e := Register(foriegnKeyValidationError{}); e != nil {
		t.Errorf("Not expecting error on TestCoverageOnNoDbFieldSkip, but got: %s", e.Error())
	}
}

func TestCanSetDbTableNameSeparately(t *testing.T) {
	defer reset()
	type anonymousStructFieldConfig struct {
		_  bool `db:"database=my_database"`
		_  bool `db:"table=my_table"`
		ID int  `db:"col=id,pk,ai"`
	}
	if e := Register(anonymousStructFieldConfig{}); e != nil {
		t.Errorf("Wasn't expecting error on TestCanSetDbTableNameSeparately, but got: %s", e.Error())
	}
	if _, ok := modMap["anonymousStructFieldConfig"]; !ok {
		t.Errorf("Expected modMap to be set for anonymousStructFieldConfig")
	}

	if _, ok := dbMap["my_database"]; !ok {
		t.Fatalf("Expected dbMap to be set for my_database")
	}

	col, ok := dbMap["my_database"]["my_table"]
	if !ok {
		t.Fatalf(`Expected dbMap["my_datgabase"] to be set for my_table`)
	}

	if len(col) != 1 {
		t.Fatalf("Expected exactly one column in table map, but got %d", len(col))
	}
}

func TestAllTagsPass(t *testing.T) {
	defer reset()
	type allTagsPass struct {
		non       bool   `non-db:"I'm not a DB tag..."`
		ForeignID string `db:"col=foreign_id"`
		ID        int    `db:"database=foo,table=bar,col=baz,fkmap=foreign_id.Model.Field,pk,ai,null"`
	}
	if e := Register(allTagsPass{}); e != nil {
		t.Errorf("Not expecting error on TestAllTagsPass but got: %s", e.Error())
	}

	if len(modMap) != 1 {
		fmt.Printf("%+v", modMap)
		t.Errorf("Expected length of modMap to be exactly 1, but got %d", len(modMap))
	}

	mod, ok := modMap["allTagsPass"]
	if !ok {
		t.Errorf("Expecting a database.table map for allTagsPass, but not registered in modMap")
	} else {
		if len(mod) != 2 {
			t.Errorf(`Expecting modMap for allTagsPass to be slice of two strings, but got: %+v`, mod)
		}
	}

	tblMap, ok := dbMap[mod[0]]
	if !ok {
		t.Fatalf(`Expected to get a table map for datbase name %s, but not registered in dbMap`, mod[0])
	}

	if len(tblMap) != 1 {
		t.Errorf(`Expected length of dbMap["%s"] to be exactly 1, but got %d`, mod[0], len(tblMap))
	}

	colSlice, ok := tblMap[mod[1]]
	if !ok {
		t.Fatalf(`Expected to get a slice of columns for db %s, table %s, but not registered in dbMap`, mod[0], mod[1])
	}

	if len(colSlice) != 2 {
		t.Fatalf(`Expected length of []columns for db %s, table %s to be exactly 2, %d found`, mod[0], mod[1], len(colSlice))
	}

	// Foreign key column defined before foreign key map
	col := colSlice[0]
	if col.fieldName != "ForeignID" {
		t.Errorf(`Expected first defined column field to be ForeignID, but got "%s"`, col.fieldName)
	}

	if col.colName != "foreign_id" {
		t.Errorf(`Expected first defined column name to be foreign_id, but got "%s"`, col.colName)
	}

	if col.colType != reflect.String {
		t.Errorf(`Expected foreign_id colum to be kind String`)
	}

	// Column with all parameters set and foreign key map to foreign_id column above
	col = colSlice[1]
	if col.fieldName != "ID" {
		t.Errorf(`Expected column fieldName to be "ID" but got "%s"`, col.fieldName)
	}

	if col.colName != "baz" {
		t.Errorf(`Expected column colName to be "baz" but got "%s"`, col.colName)
	}

	if col.colType != reflect.Int {
		t.Errorf(`Expected column colType to be reflect.Int buyt got %T`, col.colType)
	}

	if col.pk != true {
		t.Errorf(`Expected column pk to be true but got false`)
	}

	if col.ai != true {
		t.Errorf(`Expected column ai to be true but got false`)
	}

	if col.fk != true {
		t.Errorf(`Expected column fk to be true, but got false`)
	}

	if col.colRelation != "foreign_id.Model.Field" {
		t.Errorf(`Expected column colRelation to be "baz.Model.Field but got %s"`, col.colRelation)
	}

	if col.null != true {
		t.Errorf(`Expected column null to be true, but got false`)
	}
}
