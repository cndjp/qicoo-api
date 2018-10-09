package handler_test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/cndjp/qicoo-api/src/handler"
	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
)

var travisTestRedisConn redis.Conn
var internalTestRedisConn *redigomock.Conn

const testEventID = "testEventID"

var mockQuestion = handler.Question{
	ID:        "1",
	Object:    "question",
	Username:  "anonymous",
	EventID:   "0",
	ProgramID: "1",
	Comment:   "I am mock",
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
	Like:      100000,
}

var mockMuxVars = handler.MuxVars{
	EventID: testEventID,
	Start:   1,
	End:     100,
	Sort:    "created_at",
	Order:   "asc",
}

//type redigoMockConn struct {
//	conn         redis.Conn
//	redisCommand string
//	sortedkey    string
//	questionList handler.QuestionList
//	eventID      string
//}
//
//func (m redigoMockConn) GetRedisConnection() redis.Conn {
//	return m.conn
//}
//
//func (m redigoMockConn) selectRedisCommand() string {
//	return m.redisCommand
//}
//func (m redigoMockConn) selectRedisSortedKey() string {
//	return m.sortedkey
//}
//func (m redigoMockConn) GetQuestionList() handler.QuestionList {
//	return m.questionList
//}
//func (m redigoMockConn) SetQuestion(question handler.Question) error {
//	return nil
//}
//
//func (m redigoMockConn) getQuestionsKey() {
//	return
//}
//
//func (m redigoMockConn) checkRedisKey() {
//	return
//}
//
//func (m redigoMockConn) syncQuestion() string {
//	return m.eventID
//}

func isTravisEnv() bool {
	if os.Getenv("TRAVIS") == "true" {
		return true
	}
	return false
}

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	if isTravisEnv() {
		conn, err := redis.Dial("tcp", "localhost:6379")
		if err != nil {
			log.Fatal(err)
		}

		travisTestRedisConn = conn
	} else {
		conn := redigomock.NewConn()
		defer func() {
			conn.Clear()
			err := conn.Close()
			if err != nil {
				log.Fatal("runTests: failed launch redis server:", err)
			}
		}()

		internalTestRedisConn = conn
	}

	return m.Run()
}

func flushallRedis(conn redis.Conn) {
	if _, err := conn.Do("FLUSHALL"); err != nil {
		fmt.Println(err)
	}
}

func newMockPool() *handler.RedisClient {
	m := new(handler.RedisClient)
	if isTravisEnv() {
		m.RedisConn = travisTestRedisConn
	} else {
		m.RedisConn = internalTestRedisConn
	}

	m.Vars = mockMuxVars

	return m
}
