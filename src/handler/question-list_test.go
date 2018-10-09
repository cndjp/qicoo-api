package handler_test

import (
	"encoding/json"
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

func TestGetQuestionListInTheTravis(t *testing.T) {
	var mockPool = newMockPool()
	defer func() {
		mockPool.RedisConn.Close()

		// 一律でflushallはやりすぎか？
		flushallRedis(mockPool.RedisConn)
	}()
	var mockChannel = testEventID
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

// ローカルのモックでやるやつ
func TestGetQuestionListInTheLocal(t *testing.T) {
	var mockPool = newMockPool()
	defer func() {
		mockPool.RedisConn.Close()

		// 一律でflushallはやりすぎか？
		internalTestRedisConn.Command("FLUSHALL").Expect("OK")
		flushallRedis(mockPool.RedisConn)
	}()
	var mockChannel = testEventID

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
