package handler_test

import (
	"reflect"
	"testing"

	"github.com/cndjp/qicoo-api/src/handler"
	"github.com/sirupsen/logrus"
)

func judgeGetQuestionLike(q handler.Question, t *testing.T) {
	mockLike := q.Like
	expectedLike := 1

	if !reflect.DeepEqual(expectedLike, mockLike) {
		t.Errorf("expected %q to eq %q", expectedLike, mockLike)
	}
}

// TestQuestionLike
func TestQuestionLike(t *testing.T) {
	var rci handler.RedisConnectionInterface
	rci = new(mockRedisManager)

	var dmi handler.MySQLDbmapInterface
	dmi = new(mockMySQLManager)

	var q handler.Question
	q, err := handler.QuestionLikeFunc(rci, dmi, mockQLikeMuxVars)

	if err != nil {
		logrus.Error(err)
		t.Errorf("error :%v", err)
	}

	judgeGetQuestionLike(q, t)
}
