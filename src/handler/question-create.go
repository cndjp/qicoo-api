package handler

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/cndjp/qicoo-api/src/loglib"
	"github.com/go-gorp/gorp"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// QuestionCreateHandler QuestionオブジェクトをDBとRedisに書き込む
func QuestionCreateHandler(w http.ResponseWriter, r *http.Request) {
	sugar := loglib.GetSugar()
	defer sugar.Sync()

	// Headerの設定
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// DBとRedisに書き込むためのstiruct Object を生成。POST REQUEST のBodyから値を取得
	var question Question
	decoder := json.NewDecoder(r.Body)

	var err error
	err = decoder.Decode(&question)
	if err != nil {
		sugar.Error(err)
		return
	}

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

	sugar.Infof("Request QuestionCreate process. EventID:%s, QuestionID:%s, Comment:%s, Username:%s, ProgramID:%s, Like:%d, CreatedAt:%s", v.EventID, question.ID, question.Comment, question.Username, question.ProgramID, question.Like, question.CreatedAt)

	var rci RedisConnectionInterface
	rci = new(RedisManager)

	var dmi MySQLDbmapInterface
	dmi = new(MySQLManager)

	err = QuestionCreateFunc(rci, dmi, v, question)
	if err != nil {
		sugar.Error(err)
		return
	}

	/* Response JSONの整形 */
	// QuestionのStructをjsonとして変換
	jsonBytes, err := json.Marshal(question)
	if err != nil {
		sugar.Error(err)
		return
	}

	// 整形用のバッファを作成し、整形を実行
	out := new(bytes.Buffer)
	// プリフィックスなし、スペース2つでインデント
	err = json.Indent(out, jsonBytes, "", "  ")
	if err != nil {
		sugar.Error(err)
		return
	}

	w.Write([]byte(out.String()))
	sugar.Infof("Response QuestionCreate process. QuestionCreate:%s", jsonBytes)
}

// QuestionCreateFunc テストコードでテストしやすいように定義
func QuestionCreateFunc(rci RedisConnectionInterface, dmi MySQLDbmapInterface, v QuestionCreateMuxVars, question Question) error {

	var dbmap *gorp.DbMap
	dbmap = dmi.GetMySQLdbmap()
	defer dbmap.Db.Close()

	// gorpのトランザクション処理。DBとRedisの両方とも書き込みが出来た場合に、commitする
	trans, err := dbmap.Begin()
	if err != nil {
		return err
	}

	err = QuestionCreateDB(dbmap, question)
	if err != nil {
		trans.Rollback()
		return err
	}

	err = QuestionCreateRedis(rci, dmi, v, question)
	if err != nil {
		trans.Rollback()
		return err
	}

	err = trans.Commit()
	if err != nil {
		trans.Rollback()
		return err
	}

	return nil
}

// QuestionCreateDB DBに質問データの挿入
func QuestionCreateDB(dbmap *gorp.DbMap, question Question) error {

	// debug SQL Trace
	dbmap.TraceOn("", log.New(os.Stdout, "gorptrace: ", log.LstdFlags))

	/* データの挿入 */
	return dbmap.Insert(&question)
}

// SetQuestion QuestionをRedisに格納
func SetQuestion(redisConn redis.Conn, dmi MySQLDbmapInterface, v QuestionCreateMuxVars, question Question) error {
	sugar := loglib.GetSugar()
	defer sugar.Sync()

	// RedisClient にKeyを生成
	rks := GetRedisKeys(v.EventID)

	// 多分並列処理できるやつ
	/* Redisにデータが存在するか確認する。 */
	yes, err := checkRedisKey(redisConn, rks)
	if err != nil {
		return err
	}

	if !yes {
		dbmap := dmi.GetMySQLdbmap()
		defer dbmap.Db.Close()
		_, err := syncQuestion(redisConn, dbmap, v.EventID, rks)
		// 同期にエラー
		if err != nil {
			return err
		}
	}

	//HashMap SerializedされたJSONデータを格納
	serializedJSON, err := json.Marshal(question)
	if err != nil {
		return err
	}

	sugar.Infof("Redis Command of SetQuestion. command='HSET %s %s %s'", rks.QuestionKey, question.ID, serializedJSON)
	if _, err := redisConn.Do("HSET", rks.QuestionKey, question.ID, serializedJSON); err != nil {
		return err
	}

	//SortedSet(Like)
	sugar.Infof("Redis Command of SetQuestion. command='ZADD %s %d %s'", rks.LikeSortedKey, question.Like, question.ID)
	if _, err := redisConn.Do("ZADD", rks.LikeSortedKey, question.Like, question.ID); err != nil {
		return err
	}

	//SortedSet(CreatedAt)
	sugar.Infof("Redis Command of SetQuestion. command='ZADD %s %d %s'", rks.CreatedSortedKey, question.CreatedAt.Unix(), question.ID)
	if _, err := redisConn.Do("ZADD", rks.CreatedSortedKey, question.CreatedAt.Unix(), question.ID); err != nil {
		return err
	}

	return nil
}

// QuestionCreateRedis Redisに質問データの挿入
func QuestionCreateRedis(rci RedisConnectionInterface, dmi MySQLDbmapInterface, v QuestionCreateMuxVars, question Question) error {

	// RedisのConnection生成
	redisConn := rci.GetRedisConnection()
	defer redisConn.Close()

	return SetQuestion(redisConn, dmi, v, question)
}
