package handler_test

import (
	"reflect"
	"testing"

	"github.com/cndjp/qicoo-api/src/handler"
	"github.com/sirupsen/logrus"
)

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
	err := handler.QuestionCreateFunc(rci, dmi, mockQCMuxVars, testQuestion)
	if err != nil {
		logrus.Error(err)
		t.Errorf("error in saving question :%v", err)
	}

	// QuestionListを取得し、testQuestionが含まれているか確認
	var questionList handler.QuestionList
	questionList, err = handler.QuestionListFunc(rci, dmi, mockQLMuxVars)

	if err != nil {
		logrus.Error(err)
		t.Errorf("error :%v", err)
	}

	judgeCreateQuestionList(questionList, t)
}
