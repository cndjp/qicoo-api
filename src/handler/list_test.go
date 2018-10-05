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

type redigoMockConn struct {
	conn redis.Conn
}

func (m redigoMockConn) GetRedisConnection() redis.Conn {
	return m.conn
}

func (m redigoMockConn) Close() error {
	return m.conn.Close()
}

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

	internalTestRedisConn = conn

	return m.Run()
}

func flushallRedis(conn redis.Conn) {
	if _, err := conn.Do("FLUSHALL"); err != nil {
		fmt.Println(err)
	}
}

func TestGetQuestionListInTheTravis(t *testing.T) {
	localConn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		t.Error(err)
	}

	mockP := &redigoMockConn{
		conn: localConn,
	}

	var pool = &redis.Pool{
		MaxIdle:     3,
		MaxActive:   1000,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return localConn, nil },
	}

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

	var mockPool = &handler.RedisPool{
		PIface: mockP,
		Pool:   pool,
		Vars: handler.MuxVars{
			EventID: testEventID,
			Start:   1,
			End:     100,
			Sort:    "created_at",
			Order:   "asc",
		},
	}

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

	ql := mockPool.GetQuestionList()

	mockComment := ql.Data[0].Comment
	expectedComment := "I am mock"

	if !reflect.DeepEqual(expectedComment, mockComment) {
		t.Errorf("expected %q to eq %q", expectedComment, mockComment)
	}
}

//mockからプール読んでくる処理が無理っぽい
func TestGetQuestionListInTheLocal(t *testing.T) {
	mockP := &redigoMockConn{
		conn: internalTestRedisConn,
	}

	mockRP := handler.RedisPool{
		PIface: mockP,
		Vars: handler.MuxVars{
			EventID: testEventID,
			Start:   1,
			End:     100,
			Sort:    "created_at",
			Order:   "asc",
		},
	}

	mockRP.RedisConn = mockRP.GetInterfaceRedisConnection()
	defer func() {
		mockRP.RedisConn.Close()

		// 一律でflushallはやりすぎか？
		flushallRedis(mockRP.RedisConn)
	}()

	internalTestRedisConn.Command("FLUSHALL").Expect("OK")
	defer flushallRedis(internalTestRedisConn)

	var mockQuestion = handler.Question{
		ID:        "1",
		Object:    "question",
		Username:  "anonymous",
		EventID:   "0",
		ProgramID: "0",
		Comment:   "I am mock",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Like:      100000,
	}

	var mockChannel = testEventID

	mockQuestionJS, _ := json.Marshal(mockQuestion)

	internalTestRedisConn.Command("HMGET", "questions_"+mockChannel, "1").ExpectSlice(mockQuestionJS, nil)
	internalTestRedisConn.Command("HSET", "questions_"+mockChannel, 1, mockQuestionJS)                                         //.Expect(int64(1))
	internalTestRedisConn.Command("ZADD", "questions_"+mockChannel+"_like", mockQuestion.Like, mockQuestion.ID)                //.Expect(int64(1))
	internalTestRedisConn.Command("ZADD", "questions_"+mockChannel+"_created", mockQuestion.CreatedAt.Unix(), mockQuestion.ID) //.Expect(int64(1))
	internalTestRedisConn.Command("EXISTS", "questions_"+mockChannel)                                                          //.Expect(int64(1))
	internalTestRedisConn.Command("EXISTS", "questions_"+mockChannel+"_like")                                                  //.Expect(int64(1))
	internalTestRedisConn.Command("EXISTS", "questions_"+mockChannel+"_created")                                               //.Expect(int64(1))
	internalTestRedisConn.Command("ZRANGE", "questions_"+mockChannel, 0, 99).Expect([]interface{}{
		mockQuestion,
	})
	internalTestRedisConn.Command("ZRANGE", "questions_"+mockChannel+"_created", 0, 99).Expect([]interface{}{
		"1",
	})

	if _, err := internalTestRedisConn.Do("HSET", "questions_"+mockChannel, 1, mockQuestionJS); err != nil {
		t.Error(err)
	}

	internalTestRedisConn.Command("HGET", "questions_"+mockChannel, 1).Expect(int64(1))
	fmt.Println(internalTestRedisConn.Do("HGET", "questions_"+mockChannel, 1))

	//SortedSet(Like)
	if _, err := internalTestRedisConn.Do("ZADD", "questions_"+mockChannel+"_like", mockQuestion.Like, mockQuestion.ID); err != nil {
		t.Error(err)
	}

	//SortedSet(CreatedAt)
	if _, err := internalTestRedisConn.Do("ZADD", "questions_"+mockChannel+"_created", mockQuestion.CreatedAt.Unix(), mockQuestion.ID); err != nil {
		t.Error(err)
	}

	ql := mockRP.GetQuestionList()

	mockComment := ql.Data[0].Comment
	expectedComment := "I am mock"

	if !reflect.DeepEqual(expectedComment, mockComment) {
		t.Errorf("expected %q to eq %q", expectedComment, mockComment)
	}
}
