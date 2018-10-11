package handler

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	_ "time"

	"github.com/go-gorp/gorp"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// QuestionCreateHandler QuestionオブジェクトをDBとRedisに書き込む
func QuestionCreateHandler(w http.ResponseWriter, r *http.Request) {

	// DBとRedisに書き込むためのstiruct Object を生成。POST REQUEST のBodyから値を取得
	var question Question
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&question)

	// POST REQUEST の BODY に含まれていない値の設定
	question.ID = uuid.New().String()
	question.Object = "question"
	question.Username = "anonymous"
	question.EventID = mux.Vars(r)["event_id"]
	question.ProgramID = "1" // 未実装のため1固定で生成
	question.Like = 0
	question.UpdatedAt = TimeNowRoundDown()
	question.CreatedAt = question.UpdatedAt

	// URLに含まれている event_id を取得
	vars := mux.Vars(r)
	v := QuestionCreateMuxVars{
		EventID: vars["event_id"],
	}

	var rci RedisConnectionInterface
	rci = new(RedisManager)

	var dmi MySQLDbmapInterface
	dmi = new(MySQLManager)

	err := QuestionCreateFunc(rci, dmi, v, question)
	if err != nil {
		logrus.Error(err)
		return
	}

	/* Response JSONの整形 */
	// QuestionのStructをjsonとして変換
	jsonBytes, err := json.Marshal(question)
	if err != nil {
		logrus.Error(err)
		return
	}

	// 整形用のバッファを作成し、整形を実行
	out := new(bytes.Buffer)
	// プリフィックスなし、スペース2つでインデント
	json.Indent(out, jsonBytes, "", "  ")

	w.Write([]byte(out.String()))
}

// QuestionCreateFunc テストコードでテストしやすいように定義
func QuestionCreateFunc(rci RedisConnectionInterface, dmi MySQLDbmapInterface, v QuestionCreateMuxVars, question Question) error {
	var dbmap *gorp.DbMap
	dbmap = dmi.GetMySQLdbmap()
	defer dbmap.Db.Close()

	// gorpのトランザクション処理。DBとRedisの両方とも書き込みが出来た場合に、commitする
	trans, err := dbmap.Begin()
	if err != nil {
		logrus.Error(err)
		return err
	}

	err = QuestionCreateDB(dbmap, question)
	if err != nil {
		logrus.Error(err)
		trans.Rollback()
		return err
	}

	err = QuestionCreateRedis(rci, dmi, v, question)
	if err != nil {
		logrus.Error(err)
		trans.Rollback()
		return err
	}

	trans.Commit()
	return nil
}

// QuestionCreateDB DBに質問データの挿入
func QuestionCreateDB(dbmap *gorp.DbMap, question Question) error {
	// debug SQL Trace
	dbmap.TraceOn("", log.New(os.Stdout, "gorptest: ", log.Lmicroseconds))

	/* データの挿入 */
	err := dbmap.Insert(&question)

	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

// SetQuestion QuestionをRedisに格納
func SetQuestion(redisConn redis.Conn, dmi MySQLDbmapInterface, v QuestionCreateMuxVars, question Question) error {
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

	//HashMap SerializedされたJSONデータを格納
	serializedJSON, err := json.Marshal(question)
	if err != nil {
		logrus.Error(err)
		return err
	}

	if _, err := redisConn.Do("HSET", rks.QuestionKey, question.ID, serializedJSON); err != nil {
		logrus.Error(err)
		return err
	}

	//SortedSet(Like)
	if _, err := redisConn.Do("ZADD", rks.LikeSortedKey, question.Like, question.ID); err != nil {
		logrus.Error(err)
		return err
	}

	//SortedSet(CreatedAt)
	if _, err := redisConn.Do("ZADD", rks.CreatedSortedKey, question.CreatedAt.Unix(), question.ID); err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

// QuestionCreateRedis Redisに質問データの挿入
func QuestionCreateRedis(rci RedisConnectionInterface, dmi MySQLDbmapInterface, v QuestionCreateMuxVars, question Question) error {
	// RedisのConnection生成
	redisConn := rci.GetRedisConnection()
	defer redisConn.Close()

	err := SetQuestion(redisConn, dmi, v, question)

	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}
