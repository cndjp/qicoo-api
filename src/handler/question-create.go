package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

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
	question.UpdatedAt = time.Now()
	question.CreatedAt = time.Now()

	// debug
	w.Write([]byte("comment: " + question.Comment + "\n" +
		"ID: " + question.ID + "\n" +
		"Object: " + question.Object + "\n" +
		"eventID: " + question.EventID + "\n" +
		"programID: " + question.ProgramID + "\n" +
		"username: " + question.Username + "\n" +
		"Like: " + strconv.Itoa(question.Like) + "\n"))

	var dmi MySQLDbmapInterface
	dmi = new(MySQLManager)

	CreateQuestionDB(dmi, question)

	var rci RedisConnectionInterface
	rci = new(RedisManager)

	// URLに含まれている event_id を取得
	vars := mux.Vars(r)
	v := QuestionCreateMuxVars{
		EventID: vars["event_id"],
	}

	CreateQuestionRedis(rci, v, question)
}

// CreateQuestionDB DBに質問データの挿入
func CreateQuestionDB(dmi MySQLDbmapInterface, question Question) error {
	var dbmap *gorp.DbMap
	dbmap = dmi.GetMySQLdbmap()
	defer dbmap.Db.Close()

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
func SetQuestion(redisConn redis.Conn, v QuestionCreateMuxVars, question Question) error {
	// RedisClient にKeyを生成
	rks := GetRedisKeys(v.EventID)
	// TODO
	//checkRedisKey()

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

// CreateQuestionRedis Redisに質問データの挿入
func CreateQuestionRedis(rci RedisConnectionInterface, v QuestionCreateMuxVars, question Question) error {
	// RedisのConnection生成
	redisConn := rci.GetRedisConnection()
	defer redisConn.Close()

	err := SetQuestion(redisConn, v, question)

	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}
