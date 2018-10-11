package handler_test

import (
	_ "encoding/json"
	"reflect"
	"testing"

	"github.com/cndjp/qicoo-api/src/handler"
)

func judgeGetQuestionList(ql handler.QuestionList, t *testing.T) {
	mockComment := ql.Data[0].Comment
	expectedComment := "I am Mock"

	if !reflect.DeepEqual(expectedComment, mockComment) {
		t.Errorf("expected %q to eq %q", expectedComment, mockComment)
	}
}

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
