package handler_test

import (
	"reflect"
	"testing"

	"github.com/cndjp/qicoo-api/src/handler"
	"github.com/sirupsen/logrus"
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
	questionList, err := handler.QuestionListFunc(rci, dmi, mockQLMuxVars)

	if err != nil {
		logrus.Error(err)
		t.Errorf("error :%v", err)
	}

	judgeGetQuestionList(questionList, t)
}
