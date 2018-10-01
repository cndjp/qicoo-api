package sql

import (
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestInitMySQLDB(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	m := MappingDBandTable(db)
	err = m.TruncateTables()
	if err != nil {
		t.Fatalf("error truncate", err)
	}
}
