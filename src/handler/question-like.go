package handler

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cndjp/qicoo-api/src/loglib"
	"github.com/go-gorp/gorp"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
)

// QuestionLikeHandler Questionオブジェクトにいいね！をカウントアップする
func QuestionLikeHandler(w http.ResponseWriter, r *http.Request) {
	sugar := loglib.GetSugar()
	defer sugar.Sync()

	// Headerの設定
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// URLに含まれているパラメータを取得
	vars := mux.Vars(r)

	v := QuestionLikeMuxVars{
		EventID:    vars["event_id"],
		QuestionID: vars["question_id"],
	}

	sugar.Infof("Request QuestionLike process. EventID:%s, QuestionID:%s", v.EventID, v.QuestionID)

	var dmi MySQLDbmapInterface = new(MySQLManager)
	var rci RedisConnectionInterface = new(RedisManager)

	q, err := QuestionLikeFunc(rci, dmi, v)
	if err != nil {
		sugar.Error(err)
		return
	}

	/* Response JSONの整形 */
	// QuestionのStructをjsonとして変換
	jsonBytes, err := json.Marshal(q)
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

	sugar.Infof("Response QuestionLike process. QuestionLike:%s", jsonBytes)
}

// QuestionLikeFunc テストコードでテストしやすいように定義
func QuestionLikeFunc(rci RedisConnectionInterface, dmi MySQLDbmapInterface, v QuestionLikeMuxVars) (Question, error) {
	sugar := loglib.GetSugar()
	defer sugar.Sync()

	var dbmap = dmi.GetMySQLdbmap()
	//defer dbmap.Db.Close()
	dbmap.TraceOn("", log.New(os.Stdout, "gorptrace: ", log.LstdFlags))

	var question Question

	redisConn := rci.GetRedisConnection()
	defer redisConn.Close()

	// Keyを生成
	rks := GetRedisKeys(v.EventID)

	// Redisにデータが無ければ、DBと同期する。(DBへLikeをUpdateするためには、questionオブジェクトを取得する必要があり、Redisから取得する)
	yes, err := checkRedisKey(redisConn, rks)
	if err != nil {
		return question, err
	}

	if !yes {
		_, err := syncQuestion(redisConn, dbmap, v.EventID, rks)
		// 同期にエラー
		if err != nil {
			return question, err
		}
	}

	// Redisから更新対象のjsonデータを取得
	likeQuestion, err := getQuestionRedis(redisConn, v, rks)
	if err != nil {
		return question, err
	}

	// gorpのトランザクション処理。DBとRedisの両方とも書き込みが出来た場合に、commitする
	trans, err := dbmap.Begin()
	if err != nil {
		trans.Rollback()
		return question, err
	}

	err = QuestionLikeDB(dbmap, v, likeQuestion)
	if err != nil {
		trans.Rollback()
		return question, err
	}

	question, err = QuestionLikeRedis(redisConn, v, rks)
	if err != nil {
		trans.Rollback()
		return question, err
	}

	trans.Commit()
	return question, nil
}

// QuestionLikeDB MySQL上でLikeを増やす
func QuestionLikeDB(m *gorp.DbMap, v QuestionLikeMuxVars, question Question) error {
	sugar := loglib.GetSugar()
	defer sugar.Sync()

	sugar.Infof("SQL of QuestionLikeDB. SQL='UPDATE questions SET like_count=like_count+1 WHERE id = %s'", v.QuestionID)
	//stmtUpd, err := m.Prepare(fmt.Sprintf("UPDATE questions SET like_count=like_count+1 WHERE id = ?"))
	//defer stmtUpd.Close()
	//_, err = stmtUpd.Exec(v.QuestionID)

	question.Like = question.Like + 1
	_, err := m.Update(&question)

	return err
}

// QuestionLikeRedis RedisでLikeを増やす
func QuestionLikeRedis(conn redis.Conn, v QuestionLikeMuxVars, rks RedisKeys) (Question, error) {
	sugar := loglib.GetSugar()
	defer sugar.Sync()

	var q Question

	//HashからQuesitonのデータを取得する
	sugar.Infof("Redis Command of QuestionLikeRedis. command='HMGET %s %s'", rks.QuestionKey, v.QuestionID)
	bytesSlice, err := redis.ByteSlices(conn.Do("HMGET", rks.QuestionKey, v.QuestionID))
	if err != nil {
		return q, err
	}

	for _, bytes := range bytesSlice {
		err = json.Unmarshal(bytes, &q)
		if err != nil {
			return q, err
		}
	}

	//Likeのカウントアップ
	q.Like = q.Like + 1

	//再度Hash格納し直す
	serializedJSON, err := json.Marshal(q)
	if err != nil {
		return q, err
	}

	sugar.Infof("Redis Command of QuestionLikeRedis. command='HSET %s %s %s'", rks.QuestionKey, q.ID, serializedJSON)
	if _, err := conn.Do("HSET", rks.QuestionKey, q.ID, serializedJSON); err != nil {
		return q, err
	}

	//SortedKeyは既に存在しているvalueに対して新たなscore(likeの数)でZADDを実施すると正しく上書きが出来る
	sugar.Infof("Redis Command of QuestionLikeRedis. command='ZADD %s %d %s'", rks.LikeSortedKey, q.Like, q.ID)
	if _, err := conn.Do("ZADD", rks.LikeSortedKey, q.Like, q.ID); err != nil {
		return q, err
	}

	return q, nil
}

// getQuestionRedis Redisから単体の質問を取得する。
func getQuestionRedis(conn redis.Conn, v QuestionLikeMuxVars, rks RedisKeys) (question Question, err error) {
	q := new(Question)

	bytesSlice, err := redis.ByteSlices(conn.Do("HMGET", rks.QuestionKey, v.QuestionID))

	if err != nil {
		return *q, err
	}

	err = json.Unmarshal(bytesSlice[0], q)
	if err != nil {
		return *q, err
	}

	// Redis から取得したデータのtimezoneをAsia/Tokyoと指定
	locationTokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return *q, err
	}

	q.CreatedAt = q.CreatedAt.In(locationTokyo)
	q.UpdatedAt = q.UpdatedAt.In(locationTokyo)

	return *q, nil
}
