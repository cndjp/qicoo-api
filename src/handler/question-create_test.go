package handler_test

import (
	_ "encoding/json"
	_ "log"
	_ "os"
	"reflect"
	"testing"
	_ "time"

	"github.com/cndjp/qicoo-api/src/handler"
	_ "github.com/gomodule/redigo/redis"
	_ "github.com/rafaeljusto/redigomock"
)

func TestCreateQuestionRedisInTheTravis(t *testing.T) {
	//	var mockPool = newMockPool()
	//	defer func() {
	//		mockPool.RedisConn.Close()
	//
	//		// 一律でflushallはやりすぎか？
	//		flushallRedis(mockPool.RedisConn)
	//	}()
	//
	//	err := handler.CreateQuestionRedis(mockPool, mockQuestion)
	//
	//	if err != nil {
	//		t.Fatal("create question error:", err)
	//	}
}

func judgeCreateQuestionList(ql handler.QuestionList, t *testing.T) {
	mockComment := ql.Data[0].Comment
	expectedComment := "I am Test"

	if !reflect.DeepEqual(expectedComment, mockComment) {
		t.Errorf("expected %q to eq %q", expectedComment, mockComment)
	}
}

func TestCreateQuestion(t *testing.T) {
	var testQuestion handler.Question
	testQuestion.ID = "11111111-0000-0000-0000-000000000000"
	testQuestion.Object = "question"
	testQuestion.Username = "anonymous"
	testQuestion.EventID = testEventID
	testQuestion.ProgramID = "1" // 未実装のため1固定で生成
	testQuestion.Comment = "I am Test"
	testQuestion.Like = 0
	testQuestion.UpdatedAt = handler.TimeNowRoundDown()
	testQuestion.CreatedAt = testQuestion.UpdatedAt

	var rci handler.RedisConnectionInterface
	rci = new(mockRedisManager)

	var dmi handler.MySQLDbmapInterface
	dmi = new(mockMySQLManager)

	// testQuestionを実際に格納
	handler.QuestionCreateFunc(rci, dmi, mockQCMuxVars, testQuestion)

	// QuestionListを取得し、testQuestionが含まれているか確認
	var questionList handler.QuestionList
	questionList = handler.QuestionListFunc(rci, dmi, mockQLMuxVars)

	judgeCreateQuestionList(questionList, t)
}
