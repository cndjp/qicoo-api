package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "log"
	"net/http"
	"os"
	_ "strconv"
	"time"

	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Question Questionオブジェクトを扱うためのstruct
type Question struct {
	ID        string    `json:"id" db:"id"`
	Object    string    `json:"object" db:"object"`
	Username  string    `json:"username" db:"username"`
	EventID   string    `json:"event_id" db:"event_id"`
	ProgramID string    `json:"program_id" db:"program_id"`
	Comment   string    `json:"comment" db:"comment"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Like      int       `json:"like" db:"like_count"`
}

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
	//w.Write([]byte("comment: " + question.Comment + "\n" +
	//	"ID: " + question.ID + "\n" +
	//	"Object: " + question.Object + "\n" +
	//	"eventID: " + question.EventID + "\n" +
	//	"programID: " + question.ProgramID + "\n" +
	//	"username: " + question.Username + "\n" +
	//	"Like: " + strconv.Itoa(question.Like) + "\n"))

	dbmap, err := initDb()

	// debug SQL Trace
	//dbmap.TraceOn("", log.New(os.Stdout, "gorptest: ", log.Lmicroseconds))

	if err != nil {
		causeErr := errors.Cause(err)
		fmt.Printf("%+v", causeErr)
		return
	}

	/* データの挿入 */
	err = dbmap.Insert(&question)

	if err != nil {
		fmt.Printf("%+v", err)
		return
	}

	//var buf bytes.Buffer
	//enc := json.NewEncoder(&buf)
	//if err := enc.Encode(question); err != nil {
	//	log.Fatal(err)
	//}

	//fmt.Println(buf.String())

	defer dbmap.Db.Close()
}

// QuestionListHandler QuestionオブジェクトをRedisから取得する。存在しない場合はDBから取得し、Redisへ格納する
// TODO: pagenationなどのパラメータ制御。まだ仮実装
func QuestionListHandler(w http.ResponseWriter, r *http.Request) {
	// URLに含まれている event_id を取得して、questionオブジェクトに格納
	//vars := mux.Vars(r)
	//eventID := vars["event_id"]

	dbmap, err := initDb()

	if err != nil {
		causeErr := errors.Cause(err)
		fmt.Printf("%+v", causeErr)
		return
	}

	var questions []Question
	_, err = dbmap.Select(&questions, "select * from questions")

	if err != nil {
		causeErr := errors.Cause(err)
		fmt.Printf("%+v", causeErr)
		return
	}

	/* JSONの整形 */
	// QuestionのStructをjsonとして変換
	jsonBytes, err := json.Marshal(questions)

	// 整形用のバッファを作成し、整形を実行
	out := new(bytes.Buffer)
	// プリフィックスなし、スペース4つでインデント
	json.Indent(out, jsonBytes, "", "  ")

	w.Write([]byte(out.String()))

	defer dbmap.Db.Close()
}

// DbConnect 環境変数を利用しDBへのConnectionを取得する
func initDb() (dbmap *gorp.DbMap, err error) {
	dbms := "mysql"
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	protocol := "tcp(" + os.Getenv("DB_URL") + ")"
	dbname := "qicoo"
	option := "?parseTime=true"

	connect := user + ":" + password + "@" + protocol + "/" + dbname + option
	db, err := sql.Open(dbms, connect)

	if err != nil {
		return nil, errors.Wrap(err, "error on initDb()")
	}

	// structの構造体とDBのTableを紐づける
	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{}}
	dbmap.AddTableWithName(Question{}, "questions")

	return dbmap, nil
}

func main() {
	r := mux.NewRouter()

	// route QuestionCreate
	r.Path("/v1/{event_id}/questions").
		Methods("POST").
		HandlerFunc(QuestionCreateHandler)

	// route QuestionList
	r.Path("/v1/{event_id}/questions").
		Methods("GET").
		HandlerFunc(QuestionListHandler)

	http.ListenAndServe(":8080", r)
}
