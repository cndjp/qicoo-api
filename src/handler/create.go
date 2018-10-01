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

	/* POST REQUEST の BODY に含まれていない値の生成 */
	// uuid
	newUUID := uuid.New()
	question.ID = newUUID.String()

	// object
	question.Object = "question"

	// username
	// TODO: Cookieからsessionidを取得して、Redisに存在する場合は、usernameを取得してquestionオブジェクトに格納する
	question.Username = "anonymous"

	// event_id URLに含まれている event_id を取得して、questionオブジェクトに格納
	vars := mux.Vars(r)
	eventID := vars["event_id"]
	question.EventID = eventID

	// いいねの数
	question.Like = 0

	// 時刻の取得
	now := time.Now()
	question.UpdatedAt = now
	question.CreatedAt = now

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
	defer m.Db.Close()

	// debug SQL Trace
	//dbmap.TraceOn("", log.New(os.Stdout, "gorptest: ", log.Lmicroseconds))
	m.TraceOn("", log.New(os.Stdout, "gorptest: ", log.Lmicroseconds))

	if err != nil {
		logrus.Error(err)
		return
	}

	/* データの挿入 */
	//err = dbmap.Insert(&question)
	err = m.Insert(&question)

	if err != nil {
		logrus.Error(err)
		return
	}

}
