package handler_test

import (
	"log"
	"os"
	"testing"
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

func flushallRedis() {
	// TODO
}
