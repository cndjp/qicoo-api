package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cndjp/qicoo-api/src/sql"
	"github.com/go-gorp/gorp"
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

	var m *gorp.DbMap
	db, err := sql.InitMySQL()
	if err != nil {
		logrus.Error(err)
		return
	}

	m = sql.MappingDBandTable(db)
	m.AddTableWithName(Question{}, "questions")
	defer m.Db.Close()

	if err != nil {
		logrus.Error(err)
		return
	}

	CreateQuestionDB(m, question)

	// RedisClientの初期化初期設定
	rc := new(RedisClient)

	// URLに含まれている event_id を取得
	vars := mux.Vars(r)
	v := new(MuxVars)
	v.EventID = vars["event_id"]
	rc.Vars = *v

	CreateQuestionRedis(rc, question)
}

// CreateQuestionDB DBに質問データの挿入
func CreateQuestionDB(m *gorp.DbMap, question Question) error {
	// debug SQL Trace
	m.TraceOn("", log.New(os.Stdout, "gorptest: ", log.Lmicroseconds))

	/* データの挿入 */
	err := m.Insert(&question)

	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

// SetQuestion QuestionをRedisに格納
func (rc *RedisClient) SetQuestion(question Question) error {
	// RedisClient にKeyを生成
	rc.getQuestionsKey()
	rc.checkRedisKey()

	//HashMap SerializedされたJSONデータを格納
	serializedJSON, err := json.Marshal(question)
	if err != nil {
		logrus.Error(err)
		return err
	}

	if _, err := rc.RedisConn.Do("HSET", rc.QuestionsKey, question.ID, serializedJSON); err != nil {
		logrus.Error(err)
		return err
	}

	//SortedSet(Like)
	if _, err := rc.RedisConn.Do("ZADD", rc.LikeSortedKey, question.Like, question.ID); err != nil {
		logrus.Error(err)
		return err
	}

	//SortedSet(CreatedAt)
	if _, err := rc.RedisConn.Do("ZADD", rc.CreatedSortedKey, question.CreatedAt.Unix(), question.ID); err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

// CreateQuestionRedis Redisに質問データの挿入
func CreateQuestionRedis(rc *RedisClient, question Question) error {
	rc.RedisConn = GetInterfaceRedisConnection(rc)

	rc.SetQuestion(question)

	return nil
}
