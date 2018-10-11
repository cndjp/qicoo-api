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

	// 削除用のQuestionを生成 (GORPで使用するため)
	var q *Question
	q = new(Question)
	q.ID = vars["question_id"]
	q.EventID = vars["event_id"]

	v := QuestionDeleteMuxVars{
		EventID: vars["event_id"],
	}

	var dmi MySQLDbmapInterface
	dmi = new(MySQLManager)

	var rci RedisConnectionInterface
	rci = new(RedisManager)

	err := QuestionDeleteFunc(rci, dmi, v, q)
	if err != nil {
		logrus.Error(err)
		return
	}
}

// QuestionDeleteFunc テストコードでテストしやすいように定義
func QuestionDeleteFunc(rci RedisConnectionInterface, dmi MySQLDbmapInterface, v QuestionDeleteMuxVars, q *Question) error {
	var dbmap *gorp.DbMap
	dbmap = dmi.GetMySQLdbmap()
	defer dbmap.Db.Close()

	// gorpのトランザクション処理。DBとRedisの両方とも削除が出来た場合に、commitする
	trans, err := dbmap.Begin()
	if err != nil {
		logrus.Error(err)
		return err
	}

	// DBからQuestionを削除
	err = QuestionDeleteDB(dbmap, q)
	if err != nil {
		logrus.Error(err)
		trans.Rollback()
		return err
	}

	// RedisからQurstionを削除
	QuestionDeleteRedis(rci, dmi, v, *q)

	trans.Commit()
	return nil
}

// QuestionDeleteDB DBからQuestionを削除する
func QuestionDeleteDB(dbmap *gorp.DbMap, q *Question) error {
	// Tracelogの設定
	dbmap.TraceOn("", log.New(os.Stdout, "gorptest: ", log.Lmicroseconds))

	// delete実行
	_, err := dbmap.Delete(q)
	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

// QuestionDeleteRedis RedisからQuestionを削除するメソッド
func QuestionDeleteRedis(rci RedisConnectionInterface, dmi MySQLDbmapInterface, v QuestionDeleteMuxVars, question Question) error {
	// RedisのConnection生成
	redisConn := rci.GetRedisConnection()
	defer redisConn.Close()

	// RedisClient にKeyを生成
	rks := GetRedisKeys(v.EventID)

	// 多分並列処理できるやつ
	/* Redisにデータが存在するか確認する。 */
	yes := checkRedisKey(redisConn, rks)
	if !yes {
		dbmap := dmi.GetMySQLdbmap()
		defer dbmap.Db.Close()
		syncQuestion(redisConn, dbmap, v.EventID, rks)
	}

	//HashMap
	println("DeleteQuestion:", "HDEL", rks.QuestionKey, question.ID)
	if _, err := redisConn.Do("HDEL", rks.QuestionKey, question.ID); err != nil {
		logrus.Error(err)
		return err
	}

	//SortedSet Created_at
	println("DeleteQuestion:", "ZREM", rks.CreatedSortedKey, question.ID)
	if _, err := redisConn.Do("ZREM", rks.CreatedSortedKey, question.ID); err != nil {
		logrus.Error(err)
		return err
	}

	//SortedSet like
	println("DeleteQuestion:", "ZREM", rks.LikeSortedKey, question.ID)
	if _, err := redisConn.Do("ZREM", rks.LikeSortedKey, question.ID); err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}
