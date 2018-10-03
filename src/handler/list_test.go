package handler_test

import (
	"log"
	"os"
	"testing"

	"github.com/cndjp/qicoo-api/src/handler"
	"github.com/rafaeljusto/redigomock"
)

var testRedisConn *redigomock.Conn

const testEventID = "testEventID"

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

	var mockPool = handler.NewRedisPool()
	mockPool = &handler.RedisPool{
		RedisConn: testRedisConn,
		Vars: handler.MuxVars{
			EventID: testEventID,
			Start:   1,
			End:     100,
			Sort:    "created_at",
			Order:   "asc",
		},
	}

	var mockChannel = testEventID

	testRedisConn.Command("EXISTS", "questions_"+mockChannel).Expect(int64(1))
	testRedisConn.Command("EXISTS", "questions_"+mockChannel+"_like").Expect(int64(1))
	testRedisConn.Command("EXISTS", "questions_"+mockChannel+"_created").Expect(int64(1))
	testRedisConn.Command("ZRANGE", "questions_"+mockChannel+"_created", 0, 99).Expect([]interface{}{
		"one",
		"two",
		"three",
	})

	mockPool.GetQuestionList()
}
