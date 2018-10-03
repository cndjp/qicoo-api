package handler_test

import (
	"encoding/json"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/cndjp/qicoo-api/src/handler"
	"github.com/gomodule/redigo/redis"
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
	var pool = &redis.Pool{
		MaxIdle:     3,
		MaxActive:   1000,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", "localhost:6379") },
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
		Pool: pool,
		Vars: handler.MuxVars{
			EventID: testEventID,
			Start:   1,
			End:     100,
			Sort:    "created_at",
			Order:   "asc",
		},
	}

	var mockChannel = testEventID
	mockPool.RedisConn = mockPool.Pool.Get()
	defer mockPool.RedisConn.Close()

	mockQuestionJS, _ := json.Marshal(mockQuestion)

	if _, err := mockPool.RedisConn.Do("HSET", "questions_"+mockChannel, 1, mockQuestionJS); err != nil {
		t.Error(err)
	}

	//SortedSet(Like)
	mockPool.RedisConn.Do("ZADD", "questions_"+mockChannel+"_like", mockQuestion.Like, mockQuestion.ID)

	//SortedSet(CreatedAt)
	mockPool.RedisConn.Do("ZADD", "questions_"+mockChannel+"_created", mockQuestion.CreatedAt.Unix(), mockQuestion.ID)

	ql := mockPool.GetQuestionList()

	mockComment := ql.Data[0].Comment
	expectedComment := "I am mock"

	if !reflect.DeepEqual(expectedComment, mockComment) {
		t.Errorf("expected %q to eq %q", expectedComment, mockComment)
	}
}

/* うまく行かないし、一回TODOにしておこう (＾＝＾)
func TestGetQuestionList2(t *testing.T) {
	defer flushallRedis()

	var pool = &redis.Pool{
		MaxIdle:     3,
		MaxActive:   1000,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return testRedisConn, nil },
	}
	var mockPool = &handler.RedisPool{
		Pool: pool,
		Vars: handler.MuxVars{
			EventID: testEventID,
			Start:   1,
			End:     100,
			Sort:    "created_at",
			Order:   "asc",
		},
	}

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

	testRedisConn.Command("HSET", "questions_"+mockChannel, 1, mockQuestion).Expect(int64(1))
	testRedisConn.Command("ZADD", "questions_"+mockChannel+"_like", mockQuestion.Like, mockQuestion.ID).Expect(int64(1))
	testRedisConn.Command("ZADD", "questions_"+mockChannel+"_created", mockQuestion.CreatedAt.Unix(), mockQuestion.ID).Expect(int64(1))
	testRedisConn.Command("EXISTS", "questions_"+mockChannel).Expect(int64(1))
	testRedisConn.Command("EXISTS", "questions_"+mockChannel+"_like").Expect(int64(1))
	testRedisConn.Command("EXISTS", "questions_"+mockChannel+"_created").Expect(int64(1))
	testRedisConn.Command("ZRANGE", "questions_"+mockChannel, 0, 99).Expect([]interface{}{
		mockQuestion,
	})
	testRedisConn.Command("ZRANGE", "questions_"+mockChannel+"_created", 0, 99).Expect([]interface{}{
		"1",
	})

	if _, err := testRedisConn.Do("HSET", "questions_"+mockChannel, 1, mockQuestionJS); err != nil {
		t.Error(err)
	}

	//SortedSet(Like)
	testRedisConn.Do("ZADD", "questions_"+mockChannel+"_like", mockQuestion.Like, mockQuestion.ID)

	//SortedSet(CreatedAt)
	testRedisConn.Do("ZADD", "questions_"+mockChannel+"_created", mockQuestion.CreatedAt.Unix(), mockQuestion.ID)

	ql := mockPool.GetQuestionList()

	// QuestionのStructをjsonとして変換
	jsonBytes, err := json.Marshal(ql)
	if err != nil {
		logrus.Error(err)
	}

	// 整形用のバッファを作成し、整形を実行
	out := new(bytes.Buffer)
	// プリフィックスなし、スペース2つでインデント
	json.Indent(out, jsonBytes, "", "  ")

	fmt.Println(out.String())
}
*/
