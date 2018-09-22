package main

import (
	_ "bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "log"
	"net/http"
	"os"
	_ "strconv"

	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Question Questionオブジェクトを扱うためのstruct
type Question struct {
	ID        string `json:"id" db:"id"`
	Object    string `json:"object" db:"object"`
	Username  string `json:"username" db:"username"`
	EventID   string `json:"event_id" db:"event_id"`
	ProgramID string `json:"program_id" db:"program_id"`
	Comment   string `json:"comment" db:"comment"`
	CreatedAt string `json:"created_at" db:"created_at"`
	UpdateAt  string `json:"update_at" db:"update_at"`
	Like      int    `json:"like" db:"like_count"`
}

// QuestionCreateHandler QuestionオブジェクトをDBとRedisに書き込む
func QuestionCreateHandler(w http.ResponseWriter, r *http.Request) {

	// DBとRedisに書き込むためのstiruct Object を生成。POST REQUEST のBodyから値を取得
	var questions Question
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&questions)

	// POST REQUEST の BODY に含まれていない値の生成
	newUUID := uuid.New()
	questions.ID = newUUID.String()

	//questions.Object = "question"

	// TODO: Cookieからsessionidを取得して、Redisに存在する場合は、usernameを取得してquestionオブジェクトに格納する
	questions.Username = "anonymous"

	// URLに含まれている event_id を取得して、questionオブジェクトに格納
	vars := mux.Vars(r)
	eventID := vars["event_id"]
	questions.EventID = eventID

	questions.Like = 0

	dbmap, err := initDb()

	if err != nil {
		fmt.Printf("%+v", err)
	}

	var questions2 []Question
	_, err = dbmap.Select(&questions2, "select * from questions")

	if err != nil {
		fmt.Printf("%+v", err)
	}

	for x, p := range questions2 {
		fmt.Printf("    %d: %v\n", x, p)
	}

	// debug
	//	w.Write([]byte("comment: " + question.Comment + "\n" +
	//		"ID: " + question.ID + "\n" +
	//		"Object: " + question.Object + "\n" +
	//		"eventID: " + question.EventID + "\n" +
	//		"programID: " + question.ProgramID + "\n" +
	//		"Like: " + strconv.Itoa(question.Like) + "\n"))

	//var buf bytes.Buffer
	//enc := json.NewEncoder(&buf)
	//if err := enc.Encode(question); err != nil {
	//	log.Fatal(err)
	//}

	//fmt.Println(buf.String())

	defer dbmap.Db.Close()
}

// DbConnect 環境変数を利用しDBへのConnectionを取得する
func initDb() (dbmap *gorp.DbMap, err error) {
	dbms := "mysql"
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	protocol := "tcp(" + os.Getenv("DB_URL") + ")"
	dbname := "qicoo"

	connect := user + ":" + password + "@" + protocol + "/" + dbname
	db, err := sql.Open(dbms, connect)

	if err != nil {
		return nil, errors.Wrap(err, "error cant open connection")
	}

	// construct a gorp DbMap
	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	// structの構造体とDBのTableを紐づける
	dbmap.AddTableWithName(Question{}, "questions")

	return dbmap, nil
}

func main() {
	r := mux.NewRouter()

	// route QuestionCreate
	r.Path("/v1/{event_id}/questions").
		Methods("POST").
		HandlerFunc(QuestionCreateHandler)

	http.ListenAndServe(":8080", r)
}
