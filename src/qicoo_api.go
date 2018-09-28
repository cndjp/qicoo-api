package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	_ "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var version string

var (
	app = kingpin.New("qicoo-api", "This application is Qicoo's Backend API")

	verbose = app.Flag("verbose", "Run verbose mode").Default("false").Short('v').Bool()
)

// QuestionList Questionを複数格納するstruck
type QuestionList struct {
	Object string     `json:"object"`
	Type   string     `json:"type"`
	Data   []Question `json:"data"`
}

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

var redisPool *redis.Pool

// QuestionCreateHandler QuestionオブジェクトをDBとRedisに書き込む
// func QuestionCreateHandler(w http.ResponseWriter, r *http.Request) {

// 	// DBとRedisに書き込むためのstiruct Object を生成。POST REQUEST のBodyから値を取得
// 	var question Question
// 	decoder := json.NewDecoder(r.Body)
// 	decoder.Decode(&question)

// 	/* POST REQUEST の BODY に含まれていない値の生成 */
// 	// uuid
// 	newUUID := uuid.New()
// 	question.ID = newUUID.String()

// 	// object
// 	question.Object = "question"

// 	// username
// 	// TODO: Cookieからsessionidを取得して、Redisに存在する場合は、usernameを取得してquestionオブジェクトに格納する
// 	question.Username = "anonymous"

// 	// event_id URLに含まれている event_id を取得して、questionオブジェクトに格納
// 	vars := mux.Vars(r)
// 	eventID := vars["event_id"]
// 	question.EventID = eventID

// 	// いいねの数
// 	question.Like = 0

// 	// 時刻の取得
// 	now := time.Now()
// 	question.UpdatedAt = now
// 	question.CreatedAt = now

// 	// debug
// 	if *verbose {
// 		w.Write([]byte("comment: " + question.Comment + "\n" +
// 			"ID: " + question.ID + "\n" +
// 			"Object: " + question.Object + "\n" +
// 			"eventID: " + question.EventID + "\n" +
// 			"programID: " + question.ProgramID + "\n" +
// 			"username: " + question.Username + "\n" +
// 			"Like: " + strconv.Itoa(question.Like) + "\n"))
// 	}

// 	dbmap, err := initDb()
// 	defer dbmap.Db.Close()

// 	// debug SQL Trace
// 	if *verbose {
// 		dbmap.TraceOn("", log.New(os.Stdout, "gorptest: ", log.Lmicroseconds))
// 	}

// 	if err != nil {
// 		causeErr := errors.Cause(err)
// 		fmt.Printf("%+v", causeErr)
// 		return
// 	}

// 	/* データの挿入 */
// 	err = dbmap.Insert(&question)

// 	if err != nil {
// 		fmt.Printf("%+v", err)
// 		return
// 	}

// }

// QuestionListHandler QuestionオブジェクトをRedisから取得する。存在しない場合はDBから取得し、Redisへ格納する
// func QuestionListHandler(w http.ResponseWriter, r *http.Request) {
// 	// URLに含まれている event_id を取得
// 	vars := mux.Vars(r)
// 	eventID := vars["event_id"]
// 	start, _ := strconv.Atoi(vars["start"])
// 	end, _ := strconv.Atoi(vars["end"])
// 	sort := vars["sort"]
// 	order := vars["order"]

// 	questionList := getQuestionList(eventID, start, end, sort, order)

// 	/* JSONの整形 */
// 	// QuestionのStructをjsonとして変換
// 	jsonBytes, _ := json.Marshal(questionList)

// 	// 整形用のバッファを作成し、整形を実行
// 	out := new(bytes.Buffer)
// 	// プリフィックスなし、スペース2つでインデント
// 	json.Indent(out, jsonBytes, "", "  ")

// 	w.Write([]byte(out.String()))

}

// getQuestions RedisとDBからデータを取得する
// func getQuestionList(eventID string, start int, end int, sort string, order string) (questionList QuestionList) {
// 	redisConn := getRedisConnection()
// 	defer redisConn.Close()

// 	/* Redisにデータが存在するか確認する。 */
// 	questionsKey, likeSortedKey, createdSortedKey := getQuestionsKey(eventID)

// 	// 3種類のKeyが存在しない場合はデータが何かしら不足しているため、データの同期を行う
// 	hasQuestionsKey := redisHasKey(redisConn, questionsKey)
// 	hasLikeSortedKey := redisHasKey(redisConn, likeSortedKey)
// 	hasCreatedSortedKey := redisHasKey(redisConn, createdSortedKey)

// 	if !hasQuestionsKey || !hasLikeSortedKey || !hasCreatedSortedKey {
// 		syncQuestion(eventID)
// 	}

// 	/* Redisからデータを取得する */
// 	// redisのcommand
// 	var redisCommand string
// 	if order == "asc" {
// 		redisCommand = "ZRANGE"
// 	} else if order == "desc" {
// 		redisCommand = "ZREVRANGE"
// 	}

// 	// sort redisのkey
// 	var sortedkey string
// 	if sort == "created_at" {
// 		sortedkey = createdSortedKey
// 	} else if sort == "like" {
// 		sortedkey = likeSortedKey
// 	}

// 	// debug
// 	fmt.Println(redisCommand, sortedkey, start-1, end-1)

// 	// API実行時に指定されたSortをRedisで実行
// 	var uuidSlice []string
// 	uuidSlice, _ = redis.Strings(redisConn.Do(redisCommand, sortedkey, start-1, end-1))
// 	fmt.Println(uuidSlice)

// 	// RedisのDo関数は、Interface型のSliceしか受け付けないため、makeで生成 (String型のSliceはコンパイルエラー)
// 	// Example) HMGET questions_jks1812 questionID questionID questionID questionID ...
// 	var list = make([]interface{}, 0, 20)
// 	list = append(list, questionsKey)
// 	for _, str := range uuidSlice {
// 		list = append(list, str)
// 	}

// 	var bytesSlice [][]byte
// 	bytesSlice, _ = redis.ByteSlices(redisConn.Do("HMGET", list...))

// 	var questions []Question
// 	for _, bytes := range bytesSlice {
// 		q := new(Question)
// 		json.Unmarshal(bytes, q)
// 		questions = append(questions, *q)
// 	}

// 	// DB or Redis から取得したデータのtimezoneをAsia/Tokyoと指定
// 	locationTokyo, _ := time.LoadLocation("Asia/Tokyo")
// 	for i := range questions {
// 		questions[i].CreatedAt = questions[i].CreatedAt.In(locationTokyo)
// 		questions[i].UpdatedAt = questions[i].UpdatedAt.In(locationTokyo)
// 	}

// 	questionList.Data = questions
// 	questionList.Object = "list"
// 	questionList.Type = "question"
// 	return questionList
// }

// syncQuestion DBとRedisのデータを同期する
// RedisのデータがTTLなどで存在していない場合にsyncQuestionを使用する
// TODO: RedisデータのTTL
func syncQuestion(eventID string) {
	redisConnection := getRedisConnection()
	defer redisConnection.Close()

	// DBからデータを取得
	dbmap, err := initDb()
	defer dbmap.Db.Close()

	if err != nil {
		causeErr := errors.Cause(err)
		fmt.Printf("%+v", causeErr)
		return
	}

	var questions []Question
	_, err = dbmap.Select(&questions, "SELECT * FROM questions WHERE event_id = '"+eventID+"'")

	if err != nil {
		causeErr := errors.Cause(err)
		fmt.Printf("%+v", causeErr)
		return
	}

	// DB or Redis から取得したデータのtimezoneをUTCからAsia/Tokyoと指定
	locationTokyo, err := time.LoadLocation("Asia/Tokyo")
	for i := range questions {
		questions[i].CreatedAt = questions[i].CreatedAt.In(locationTokyo)
		questions[i].UpdatedAt = questions[i].UpdatedAt.In(locationTokyo)
	}

	//Redisで利用するKeyを取得
	questionsKey, likeSortedKey, createdSortedKey := getQuestionsKey(eventID)

	//DBのデータをRedisに同期する。
	for _, question := range questions {
		//HashMap SerializedされたJSONデータを格納
		serializedJSON, _ := json.Marshal(question)
		fmt.Println(questionsKey, " ", question.ID, " ", serializedJSON)
		redisConnection.Do("HSET", questionsKey, question.ID, serializedJSON)

		//SortedSet(Like)
		redisConnection.Do("ZADD", likeSortedKey, question.Like, question.ID)

		//SortedSet(CreatedAt)
		redisConnection.Do("ZADD", createdSortedKey, question.CreatedAt.Unix(), question.ID)
	}
}

// getQuestionsKey Redisで使用するQuestionsを格納するkeyを取得
func getQuestionsKey(eventID string) (questionsKey string, likeSortedKey string, createdSortedKey string) {
	questionsKey = "questions_" + eventID
	likeSortedKey = questionsKey + "_like"
	createdSortedKey = questionsKey + "_created"

	return questionsKey, likeSortedKey, createdSortedKey
}

// redisHasKey
func redisHasKey(conn redis.Conn, key string) bool {
	hasInt, _ := redis.Int(conn.Do("EXISTS", key))

	var hasKey bool
	if hasInt == 1 {
		hasKey = true
	} else {
		hasKey = false
	}

	return hasKey
}

// initDb 環境変数を利用しDBへのConnectionを取得する(sqldriverでconnection poolが実装されているらしい)
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

// initRedisPool RedisConnectionPoolからconnectionを取り出す
func initRedisPool() {
	url := os.Getenv("REDIS_URL")

	// idle connection limit:3    active connection limit:1000
	pool := &redis.Pool{
		MaxIdle:     3,
		MaxActive:   1000,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", url) },
	}

	redisPool = pool
}

// getRedisConnection
func getRedisConnection() (conn redis.Conn) {
	return redisPool.Get()
}

func init() {
	app.HelpFlag.Short('h')
	app.Version(fmt.Sprint("dntk's version: ", version))
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	// TODO
	}
}

func main() {
	r := mux.NewRouter()

	// 初期設定
	initRedisPool()

	// route QuestionCreate
	r.Path("/v1/{event_id:[a-zA-Z0-9-_]+}/questions").
		Methods("POST").
		HandlerFunc(QuestionCreateHandler)

	// route QuestionList
	r.Path("/v1/{event_id:[a-zA-Z0-9-_]+}/questions").
		Methods("GET").
		Queries("start", "{start:[0-9]+}").
		Queries("end", "{end:[0-9]+}").
		Queries("sort", "{sort:[a-zA-Z0-9-_]+}").
		Queries("order", "{order:[a-zA-Z0-9-_]+}").
		HandlerFunc(QuestionListHandler)

	http.ListenAndServe(":8080", r)
}
