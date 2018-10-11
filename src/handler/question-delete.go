package handler

import (
	"log"
	"net/http"
	"os"

	"github.com/go-gorp/gorp"
	"github.com/gomodule/redigo/redis"
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
	var rci RedisConnectionInterface
	rci = new(RedisManager)

	v := QuestionDeleteMuxVars{
		EventID: vars["event_id"],
	}

	QuestionDeleteRedis(rci, v, *q)
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
func QuestionDeleteRedis(rci RedisConnectionInterface, v QuestionDeleteMuxVars, question Question) error {
	// RedisのConnection生成
	redisConn := rci.GetRedisConnection()
	defer redisConn.Close()

	err := DeleteQuestion(redisConn, v, question)

	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

// DeleteQuestion RedisからQuestionを削除するメソッド
func DeleteQuestion(conn redis.Conn, v QuestionDeleteMuxVars, question Question) error {
	// RedisClient にKeyを生成
	rks := GetRedisKeys(v.EventID)
	checkRedisKey(conn, rks)

	//HashMap
	println("DeleteQuestion:", "HDEL", rks.QuestionKey, question.ID)
	if _, err := conn.Do("HDEL", rks.QuestionKey, question.ID); err != nil {
		logrus.Error(err)
		return err
	}

	//SortedSet Created_at
	println("DeleteQuestion:", "ZREM", rks.CreatedSortedKey, question.ID)
	if _, err := conn.Do("ZREM", rks.CreatedSortedKey, question.ID); err != nil {
		logrus.Error(err)
		return err
	}

	//SortedSet like
	println("DeleteQuestion:", "ZREM", rks.LikeSortedKey, question.ID)
	if _, err := conn.Do("ZREM", rks.LikeSortedKey, question.ID); err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}
