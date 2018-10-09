package handler

import (
	"log"
	"net/http"
	"os"

	"github.com/go-gorp/gorp"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// QuestionDeleteHandler Questionの削除用 関数
func QuestionDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// URLに含まれているパラメータを取得
	vars := mux.Vars(r)

	// 削除用のQuestion strictを生成 (GORPで使用するため)
	var q *Question
	q = new(Question)
	q.ID = vars["question_id"]
	q.EventID = vars["event_id"]

	/* DB connection 取得 */
	var m *gorp.DbMap
	m = InitMySQLQuestion()
	defer m.Db.Close()

	// DBからQuestionを削除
	err := QuestionDeleteDB(m, q)
	if err != nil {
		logrus.Error(err)
		return
	}

	// RedisからQurstionを削除
	// RedisClientの初期化初期設定
	rc := new(RedisClient)
	v := new(MuxVars)
	v.EventID = vars["event_id"]
	rc.Vars = *v

	rc.RedisConn = rc.GetRedisConnection()
	QuestionDeleteRedis(rc, *q)
}

// QuestionDeleteDB DBからQuestionを削除する
func QuestionDeleteDB(m *gorp.DbMap, q *Question) error {
	// Tracelogの設定
	m.TraceOn("", log.New(os.Stdout, "gorptest: ", log.Lmicroseconds))

	// delete実行
	_, err := m.Delete(q)
	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

// QuestionDeleteRedis RedisからQuestionを削除する
func QuestionDeleteRedis(rc *RedisClient, question Question) error {
	err := rc.DeleteQuestion(question)

	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

// DeleteQuestion RedisからQuestionを削除するメソッド
func (rc *RedisClient) DeleteQuestion(question Question) error {
	// RedisClient にKeyを生成
	rc.getQuestionsKey()
	rc.checkRedisKey()

	//HashMap
	println("DeleteQuestion:", "HDEL", rc.QuestionsKey, question.ID)
	if _, err := rc.RedisConn.Do("HDEL", rc.QuestionsKey, question.ID); err != nil {
		logrus.Error(err)
		return err
	}

	//SortedSet Created_at
	println("DeleteQuestion:", "ZREM", rc.CreatedSortedKey, question.ID)
	if _, err := rc.RedisConn.Do("ZREM", rc.CreatedSortedKey, question.ID); err != nil {
		logrus.Error(err)
		return err
	}

	//SortedSet like
	println("DeleteQuestion:", "ZREM", rc.LikeSortedKey, question.ID)
	if _, err := rc.RedisConn.Do("ZREM", rc.LikeSortedKey, question.ID); err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}
