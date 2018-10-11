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

func judgeDeleteQuestionList(ql handler.QuestionList, t *testing.T) {
	expectedComment := "I am Test"

	for _, question := range ql.Data {
		if reflect.DeepEqual(expectedComment, question.Comment) {
			t.Errorf("expected %q to not eq %q", expectedComment, question.Comment)
		}
	}
}

func TestDeleteQuestion(t *testing.T) {
	// Redisにテスト用のquestionを登録する。Create用のテストコードを実行することで流用
	TestCreateQuestion(t)

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

	// testQuestionを実際に削除
	handler.QuestionDeleteFunc(rci, dmi, mockQDMuxVars, &testQuestion)

	// QuestionListを取得し、testQuestionが削除されているか確認
	var questionList handler.QuestionList
	questionList = handler.QuestionListFunc(rci, dmi, mockQLMuxVars)

	judgeDeleteQuestionList(questionList, t)
}
