package handler

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/go-gorp/gorp"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// QuestionLikeHandler Questionオブジェクトにいいね！をカウントアップする
func QuestionLikeHandler(w http.ResponseWriter, r *http.Request) {
	// URLに含まれているパラメータを取得
	vars := mux.Vars(r)

	v := QuestionLikeMuxVars{
		EventID:    vars["event_id"],
		QuestionID: vars["question_id"],
	}

	var dmi MySQLDbmapInterface
	dmi = new(MySQLManager)

	var rci RedisConnectionInterface
	rci = new(RedisManager)

	var q Question
	var err error
	q, err = QuestionLikeFunc(rci, dmi, v)
	if err != nil {
		logrus.Error(err)
		return
	}

	/* Response JSONの整形 */
	// QuestionのStructをjsonとして変換
	jsonBytes, err := json.Marshal(q)
	if err != nil {
		logrus.Error(err)
		return
	}

	// 整形用のバッファを作成し、整形を実行
	out := new(bytes.Buffer)
	// プリフィックスなし、スペース2つでインデント
	err = json.Indent(out, jsonBytes, "", "  ")
	if err != nil {
		logrus.Error(err)
		return
	}

	w.Write([]byte(out.String()))
}

// QuestionLikeFunc テストコードでテストしやすいように定義
func QuestionLikeFunc(rci RedisConnectionInterface, dmi MySQLDbmapInterface, v QuestionLikeMuxVars) (Question, error) {
	var dbmap *gorp.DbMap
	dbmap = dmi.GetMySQLdbmap()
	defer dbmap.Db.Close()

	var question Question

	// gorpのトランザクション処理。DBとRedisの両方とも書き込みが出来た場合に、commitする
	trans, err := dbmap.Begin()
	if err != nil {
		logrus.Error(err)
		return question, err
	}

	err = QuestionLikeDB(dbmap, v)
	if err != nil {
		logrus.Error(err)
		return question, err
	}

	// RedisのConnection生成
	redisConn := rci.GetRedisConnection()
	defer redisConn.Close()

	// Keyを生成
	rks := GetRedisKeys(v.EventID)

	// Redisにデータが無ければ、DBと同期して終了。データが存在する場合はRedisでもLikeをカウントアップ
	yes, err := checkRedisKey(redisConn, rks)
	if err != nil {
		logrus.Error(err)
		return question, err
	}

	if !yes {
		_, err := syncQuestion(redisConn, dbmap, v.EventID, rks)
		// 同期にエラー
		if err != nil {
			logrus.Error(err)
			return question, err
		}

		// 同期後にQuestion取得
		question, err = getQuestion(redisConn, dbmap, v.EventID, v.QuestionID, rks)
		if err != nil {
			logrus.Error(err)
			return question, err
		}
	} else {
		question, err = QuestionLikeRedis(redisConn, v, rks)
		if err != nil {
			logrus.Error(err)
			return question, err
		}
	}

	trans.Commit()
	return question, nil
}

// QuestionLikeDB MySQL上でLikeを増やす
func QuestionLikeDB(m *gorp.DbMap, v QuestionLikeMuxVars) error {
	_, err := m.Exec("UPDATE questions SET like_count=like_count+1 WHERE id = '" + v.QuestionID + "'")
	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

// QuestionLikeRedis RedisでLikeを増やす
func QuestionLikeRedis(conn redis.Conn, v QuestionLikeMuxVars, rks RedisKeys) (Question, error) {
	var q Question

	//HashからQuesitonのデータを取得する
	bytesSlice, err := redis.ByteSlices(conn.Do("HMGET", rks.QuestionKey, v.QuestionID))
	println("QuestionLikeRedis:", "HMGET", rks.QuestionKey, v.QuestionID)
	if err != nil {
		logrus.Error(err)
		return q, err
	}

	for _, bytes := range bytesSlice {
		err = json.Unmarshal(bytes, &q)
		if err != nil {
			logrus.Error(err)
			return q, err
		}
	}

	//Likeのカウントアップ
	q.Like = q.Like + 1

	//再度Hash格納し直す
	serializedJSON, err := json.Marshal(q)
	if err != nil {
		logrus.Error(err)
		return q, err
	}

	if _, err := conn.Do("HSET", rks.QuestionKey, q.ID, serializedJSON); err != nil {
		logrus.Error(err)
		return q, err
	}

	//SortedKeyは既に存在しているvalueに対して新たなscore(likeの数)でZADDを実施すると正しく上書きが出来る
	if _, err := conn.Do("ZADD", rks.LikeSortedKey, q.Like, q.ID); err != nil {
		logrus.Error(err)
		return q, err
	}

	return q, nil
}
