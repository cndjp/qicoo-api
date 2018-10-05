package handler_test

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/cndjp/qicoo-api/src/handler"
	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
)

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

type redigoMockConn struct {
	conn redis.Conn
}

func (m redigoMockConn) GetRedisConnection() redis.Conn {
	return m.conn
}

func (m redigoMockConn) Close() error {
	return m.conn.Close()
}

func isTravisEnv() bool {
	if os.Getenv("IS_TRAVISENV") == "true" {
		return true
	}
	return false
}

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {

	if !isTravisEnv() {
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

func judgeGetQuestionList(ql handler.QuestionList, t *testing.T) {
	mockComment := ql.Data[0].Comment
	expectedComment := "I am mock"

	if !reflect.DeepEqual(expectedComment, mockComment) {
		t.Errorf("expected %q to eq %q", expectedComment, mockComment)
	}
}

func TestGetQuestionListInTheTravis(t *testing.T) {
	localConn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		t.Error(err)
	}

	var mockPool = handler.NewRedisPool()
	mockPool.PIface = &redigoMockConn{
		conn: localConn,
	}
	mockPool.Vars = mockMuxVars

	var mockChannel = testEventID
	mockPool.RedisConn = mockPool.GetInterfaceRedisConnection()
	defer func() {
		mockPool.RedisConn.Close()

		// 一律でflushallはやりすぎか？
		flushallRedis(mockPool.RedisConn)
	}()

	mockQuestionJS, err := json.Marshal(mockQuestion)
	if err != nil {
		t.Error(err)
	}

	if _, err := mockPool.RedisConn.Do("HSET", "questions_"+mockChannel, 1, mockQuestionJS); err != nil {
		t.Error(err)
	}

	//SortedSet(Like)
	if _, err := mockPool.RedisConn.Do("ZADD", "questions_"+mockChannel+"_like", mockQuestion.Like, mockQuestion.ID); err != nil {
		t.Error(err)
	}

	//SortedSet(CreatedAt)
	if _, err := mockPool.RedisConn.Do("ZADD", "questions_"+mockChannel+"_created", mockQuestion.CreatedAt.Unix(), mockQuestion.ID); err != nil {
		t.Error(err)
	}

	judgeGetQuestionList(mockPool.GetQuestionList(), t)
}

//mockからプール読んでくる処理が無理っぽい
func TestGetQuestionListInTheLocal(t *testing.T) {
	var mockPool = handler.NewRedisPool()
	mockPool.PIface = &redigoMockConn{
		conn: internalTestRedisConn,
	}
	mockPool.Vars = mockMuxVars

	var mockChannel = testEventID
	mockPool.RedisConn = mockPool.GetInterfaceRedisConnection()
	defer func() {
		mockPool.RedisConn.Close()

		// 一律でflushallはやりすぎか？
		flushallRedis(mockPool.RedisConn)
	}()

	internalTestRedisConn.Command("FLUSHALL").Expect("OK")
	defer flushallRedis(internalTestRedisConn)

	mockQuestionJS, _ := json.Marshal(mockQuestion)

	internalTestRedisConn.Command("HMGET", "questions_"+mockChannel, "1").ExpectSlice(mockQuestionJS, nil)
	internalTestRedisConn.Command("HSET", "questions_"+mockChannel, 1, mockQuestionJS)
	internalTestRedisConn.Command("ZADD", "questions_"+mockChannel+"_like", mockQuestion.Like, mockQuestion.ID)
	internalTestRedisConn.Command("ZADD", "questions_"+mockChannel+"_created", mockQuestion.CreatedAt.Unix(), mockQuestion.ID)
	internalTestRedisConn.Command("EXISTS", "questions_"+mockChannel)
	internalTestRedisConn.Command("EXISTS", "questions_"+mockChannel+"_like")
	internalTestRedisConn.Command("EXISTS", "questions_"+mockChannel+"_created")
	internalTestRedisConn.Command("ZRANGE", "questions_"+mockChannel, 0, 99).Expect([]interface{}{
		mockQuestion,
	})
	internalTestRedisConn.Command("ZRANGE", "questions_"+mockChannel+"_created", 0, 99).Expect([]interface{}{
		"1",
	})

	if _, err := internalTestRedisConn.Do("HSET", "questions_"+mockChannel, 1, mockQuestionJS); err != nil {
		t.Error(err)
	}

	//SortedSet(Like)
	if _, err := internalTestRedisConn.Do("ZADD", "questions_"+mockChannel+"_like", mockQuestion.Like, mockQuestion.ID); err != nil {
		t.Error(err)
	}

	//SortedSet(CreatedAt)
	if _, err := internalTestRedisConn.Do("ZADD", "questions_"+mockChannel+"_created", mockQuestion.CreatedAt.Unix(), mockQuestion.ID); err != nil {
		t.Error(err)
	}

	judgeGetQuestionList(mockPool.GetQuestionList(), t)
}
