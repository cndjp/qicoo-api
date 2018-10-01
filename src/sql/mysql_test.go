package sql

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/lestrrat-go/test-mysqld"
)

var testMysqld *mysqltest.TestMysqld

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	mysqld, err := mysqltest.NewMysqld(nil)
	if err != nil {
		log.Fatal("runTests: failed launch mysql server:", err)
	}
	defer mysqld.Stop()

	testMysqld = mysqld

	return m.Run()
}

func truncateTables() {
	db, err := sql.Open("mysql", testMysqld.Datasource("test", "", "", 0))
	if err != nil {
		log.Fatal("db connection error:", err)
	}
	defer db.Close()

	_, err = db.Exec("TRUNCATE test.mock")
	if err != nil {
		log.Fatal("truncate table error:", err)
	}

	/*rows, err := db.Query("SHOW TABLES")
	if err != nil {
		log.Fatal("show tables error:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		err = rows.Scan(&tableName)
		if err != nil {
			log.Fatal("show table error:", err)
		}
		_, err = db.Exec("TRUNCATE " + tableName)
		if err != nil {
			log.Fatal("truncate table error:", err)
		}
	}*/
}

func TestInitMySQLDB(t *testing.T) {
	defer truncateTables()

	db, err := sql.Open("mysql", testMysqld.Datasource("test", "", "", 0))
	if err != nil {
		t.Fatal("db connection error:", err)
	}
	defer db.Close()

	databaseRow, err := db.Query(`CREATE DATABASE test`)
	if err != nil {
		t.Fatal("create databases error:", err)
	}
	defer databaseRow.Close()

	tableRow, err := db.Query(`
		CREATE TABLE test.mock (
		id         INT          NOT NULL,
		object     VARCHAR(255) NOT NULL,
		username   VARCHAR(255) NOT NULL,
		event_id   VARCHAR(255) NOT NULL,
		program_id VARCHAR(255) NOT NULL,
		comment    TEXT         NOT NULL,
		created_at DATETIME     DEFAULT NULL,
		updated_at DATETIME     DEFAULT NULL,
		like_count INT          NOT NULL)`)
	if err != nil {
		t.Fatal("create tables error:", err)
	}
	defer tableRow.Close()

	insertRow, err := db.Query("INSERT INTO test.mock VALUES (1, 'question', 'anonymous', '1', '1', 'I am mock', NULL, NULL, 100000)")
	if err != nil {
		t.Fatal("insert row error:", err)
	}
	defer insertRow.Close()

	m := MappingDBandTable(db)
	result, err := m.Exec("SELECT comment from test.mock WHERE ID = 1")
	if err != nil {
		t.Fatal("", err)
	}
	t.Log(result)

	err = m.TruncateTables()
	if err != nil {
		t.Fatalf("an error '%s' was error truncate", err)
	}
}
