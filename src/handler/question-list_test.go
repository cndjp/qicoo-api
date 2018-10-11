package handler_test

import (
	_ "encoding/json"
	"reflect"
	"testing"

	"github.com/cndjp/qicoo-api/src/handler"
)

func judgeGetQuestionList(ql handler.QuestionList, t *testing.T) {
	mockComment := ql.Data[0].Comment
	expectedComment := "I am mock"

	if !reflect.DeepEqual(expectedComment, mockComment) {
		t.Errorf("expected %q to eq %q", expectedComment, mockComment)
	}
}

//func TestGetQuestionListInTheTravis(t *testing.T) {
//	var mockPool = newMockPool()
//	defer func() {
//		mockPool.RedisConn.Close()
//
//		// 一律でflushallはやりすぎか？？
//		flushallRedis(mockPool.RedisConn)
//	}()
//	var mockChannel = testEventID
//	mockQuestionJS, err := json.Marshal(mockQuestion)
//	if err != nil {
//		t.Error(err)
//	}
//
//	if _, err := mockPool.RedisConn.Do("HSET", "questions_"+mockChannel, 1, mockQuestionJS); err != nil {
//		t.Error(err)
//	}
//
//	//SortedSet(Like)
//	if _, err := mockPool.RedisConn.Do("ZADD", "questions_"+mockChannel+"_like", mockQuestion.Like, mockQuestion.ID); err != nil {
//		t.Error(err)
//	}
//
//	//SortedSet(CreatedAt)
//	if _, err := mockPool.RedisConn.Do("ZADD", "questions_"+mockChannel+"_created", mockQuestion.CreatedAt.Unix(), mockQuestion.ID); err != nil {
//		t.Error(err)
//	}
//
//	judgeGetQuestionList(mockPool.GetQuestionList(), t)
//}

// TestGetQuestionList
func TestGetQuestionList(t *testing.T) {
	var rci handler.RedisConnectionInterface
	rci = new(mockRedisManager)

	var dmi handler.MySQLDbmapInterface
	dmi = new(mockMySQLManager)

	var questionList handler.QuestionList
	questionList = handler.QuestionListFunc(rci, dmi, mockQLMuxVars)

	judgeGetQuestionList(questionList, t)
}
