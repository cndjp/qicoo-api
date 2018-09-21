package main

import (
	_ "bytes"
	"encoding/json"
	"fmt"
	_ "log"
	"net/http"
	"os"
	_ "strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// Question Questionオブジェクトを扱うためのstruct
type Question struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Username  string `json:"username"`
	EventID   string `json:"event_id"`
	ProgramID string `json:"program_id"`
	Comment   string `json:"comment"`
	CreatedAt string `json:"created_at"`
	UpdateAt  string `json:"update_at"`
	Like      int    `json:"like"`
}

// QuestionCreateHandler QuestionオブジェクトをDBとRedisに書き込む
func QuestionCreateHandler(w http.ResponseWriter, r *http.Request) {

	// DBとRedisに書き込むためのstiruct Object を生成。POST REQUEST のBodyから値を取得
	var question Question
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&question)

	// POST REQUEST の BODY に含まれていない値の生成
	newUUID := uuid.New()
	question.ID = newUUID.String()

	question.Object = "question"

	// TODO: Cookieからsessionidを取得して、Redisに存在する場合は、usernameを取得してquestionオブジェクトに格納する

	// URLに含まれている event_id を取得して、questionオブジェクトに格納
	vars := mux.Vars(r)
	eventID := vars["event_id"]
	question.EventID = eventID

	question.Like = 0

	connection := DbConnect()

	QuestionsEx := []Question{}
	connection.Find(&QuestionsEx, "event_id=?", "jkd1812")
	fmt.Println(QuestionsEx)

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

	defer connection.Close()
}

// DbConnect 環境変数を利用しDBへのConnectionを取得する
func DbConnect() *gorm.DB {
	dbms := "mysql"
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	protocol := "tcp(" + os.Getenv("DB_URL") + ")"
	dbname := "qicoo"

	connect := user + ":" + password + "@" + protocol + "/" + dbname
	fmt.Println("dbms: " + dbms)
	fmt.Println("connect: " + connect)
	db, err := gorm.Open(dbms, connect)

	if err != nil {
		panic(err.Error())
	}
	return db
}

//func DbQuestionCreate() {
//
//}

func main() {
	r := mux.NewRouter()

	// route QuestionCreate
	r.Path("/v1/{event_id}/questions").
		Methods("POST").
		HandlerFunc(QuestionCreateHandler)

	http.ListenAndServe(":8080", r)
}
