package handler_test

import (
	"log"
	"os"
	"testing"

	"github.com/rafaeljusto/redigomock"
)

var testRedisConn *redigomock.Conn

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	conn := redigomock.NewConn()
	defer func() {
		conn.Clear()
		err := conn.Close()
		if err != nil {
			log.Fatal("runTests: failed launch redis server:", err)
		}
	}()

	testRedisConn = conn

	return m.Run()
}

func flushallRedis() {
	testRedisConn.Command("FLUSHALL").Expect("OK")
}

func TestGetQuestionList(t *testing.T) {
	defer flushallRedis()
	//多分GetQuestionListから*redis.Connを取り出して組み直さないとテストはできない。
}
