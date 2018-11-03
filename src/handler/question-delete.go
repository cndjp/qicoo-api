package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/cndjp/qicoo-api/src/loglib"
	"github.com/go-gorp/gorp"
	"github.com/gorilla/mux"
)

// QuestionDeleteResponse Questionを削除成功した時にResponseするためのstruct
type QuestionDeleteResponse struct {
	QuesitonID string `json:"id"`
	Type       string `json:"object"`
	Deleted    bool   `json:"deleted"`
}

// QuestionDeleteHandler Questionの削除用 関数
func QuestionDeleteHandler(w http.ResponseWriter, r *http.Request) {
	sugar := loglib.GetSugar()
	defer sugar.Sync()

	// Headerの設定
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

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

	sugar.Infof("Request QuestionDelete process. EventID:%s, QuestionID:%s", v.EventID, q.ID)

	var dmi MySQLDbmapInterface
	dmi = new(MySQLManager)

	var rci RedisConnectionInterface
	rci = new(RedisManager)

	err := QuestionDeleteFunc(rci, dmi, v, q)
	if err != nil {
		sugar.Error(err)
		return
	}

	var res QuestionDeleteResponse
	res.QuesitonID = q.ID
	res.Type = "question"
	res.Deleted = true

	/* Response JSONの整形 */
	// QuestionのStructをjsonとして変換
	jsonBytes, err := json.Marshal(res)
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

	sugar.Infof("Response QuestionDelete process. QuestionDelete:%s", jsonBytes)
}

// QuestionDeleteFunc テストコードでテストしやすいように定義
func QuestionDeleteFunc(rci RedisConnectionInterface, dmi MySQLDbmapInterface, v QuestionDeleteMuxVars, q *Question) error {
	sugar := loglib.GetSugar()
	defer sugar.Sync()

	var dbmap *gorp.DbMap
	dbmap = dmi.GetMySQLdbmap()
	defer dbmap.Db.Close()

	// gorpのトランザクション処理。DBとRedisの両方とも削除が出来た場合に、commitする
	trans, err := dbmap.Begin()
	if err != nil {
		sugar.Error(err)
		return err
	}

	// DBからQuestionを削除
	err = QuestionDeleteDB(dbmap, q)
	if err != nil {
		sugar.Error(err)
		trans.Rollback()
		return err
	}

	// RedisからQurstionを削除
	err = QuestionDeleteRedis(rci, dmi, v, *q)
	if err != nil {
		sugar.Error(err)
		return err
	}

	trans.Commit()
	return nil
}

// QuestionDeleteDB DBからQuestionを削除する
func QuestionDeleteDB(dbmap *gorp.DbMap, q *Question) error {
	sugar := loglib.GetSugar()
	defer sugar.Sync()

	// Tracelogの設定
	dbmap.TraceOn("", log.New(os.Stdout, "gorptrace: ", log.Lmicroseconds))

	// delete実行
	count, err := dbmap.Delete(q)
	if err != nil {
		sugar.Error(err)
		return err
	}

	if count == 0 {
		emsg := "not found delete Quesiton. Question ID :" + q.ID
		err = errors.New(emsg)
		sugar.Error(err)
		return err
	}

	return nil
}

// QuestionDeleteRedis RedisからQuestionを削除するメソッド
func QuestionDeleteRedis(rci RedisConnectionInterface, dmi MySQLDbmapInterface, v QuestionDeleteMuxVars, question Question) error {
	sugar := loglib.GetSugar()
	defer sugar.Sync()

	// RedisのConnection生成
	redisConn := rci.GetRedisConnection()
	defer redisConn.Close()

	// RedisClient にKeyを生成
	rks := GetRedisKeys(v.EventID)

	// 多分並列処理できるやつ
	/* Redisにデータが存在するか確認する。 */
	yes, err := checkRedisKey(redisConn, rks)
	if err != nil {
		sugar.Error(err)
		return err
	}

	if !yes {
		dbmap := dmi.GetMySQLdbmap()
		defer dbmap.Db.Close()
		_, err := syncQuestion(redisConn, dbmap, v.EventID, rks)
		// 同期にエラー
		if err != nil {
			sugar.Error(err)
			return err
		}
	}

	//HashMap
	sugar.Infof("Redis Command of DeleteQuestion. command='HDEL %s %s'", rks.QuestionKey, question.ID)
	if _, err := redisConn.Do("HDEL", rks.QuestionKey, question.ID); err != nil {
		sugar.Error(err)
		return err
	}

	//SortedSet Created_at
	sugar.Infof("Redis Command of DeleteQuestion. command='ZREM %s %s'", rks.CreatedSortedKey, question.ID)
	if _, err := redisConn.Do("ZREM", rks.CreatedSortedKey, question.ID); err != nil {
		sugar.Error(err)
		return err
	}

	//SortedSet like
	sugar.Infof("Redis Command of DeleteQuestion. command='ZREM %s %s'", rks.LikeSortedKey, question.ID)
	if _, err := redisConn.Do("ZREM", rks.LikeSortedKey, question.ID); err != nil {
		sugar.Error(err)
		return err
	}

	return nil
}
