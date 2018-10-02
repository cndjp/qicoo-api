package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cndjp/qicoo-api/src/sql"
	"github.com/go-gorp/gorp"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

type RedisPool struct {
	Pool *redis.Pool
}

// NewRedisPool RedisConnectionPoolからconnectionを取り出す
func NewRedisPool() *RedisPool {
	url := os.Getenv("REDIS_URL")

	// idle connection limit:3    active connection limit:1000
	return &RedisPool{
		Pool: &redis.Pool{
			MaxIdle:     3,
			MaxActive:   1000,
			IdleTimeout: 240 * time.Second,
			Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", url) },
		},
	}

	//RedisPool = pool

}

// GetRedisConnection
func (p *RedisPool) getRedisConnection() (conn redis.Conn) {
	return p.Pool.Get()
}

// QuestionListHandler QuestionオブジェクトをRedisから取得する。存在しない場合はDBから取得し、Redisへ格納する
func (p *RedisPool) QuestionListHandler(w http.ResponseWriter, r *http.Request) {
	// URLに含まれている event_id を取得
	vars := mux.Vars(r)
	eventID := vars["event_id"]
	start, err := strconv.Atoi(vars["start"])
	if err != nil {
		logrus.Error(err)
	}
	end, err := strconv.Atoi(vars["end"])
	if err != nil {
		logrus.Error(err)
	}
	sort := vars["sort"]
	order := vars["order"]

	questionList := p.getQuestionList(eventID, start, end, sort, order)

	/* JSONの整形 */
	// QuestionのStructをjsonとして変換
	jsonBytes, err := json.Marshal(questionList)
	if err != nil {
		logrus.Error(err)
	}

	// 整形用のバッファを作成し、整形を実行
	out := new(bytes.Buffer)
	// プリフィックスなし、スペース2つでインデント
	json.Indent(out, jsonBytes, "", "  ")

	w.Write([]byte(out.String()))

}

// getQuestions RedisとDBからデータを取得する
func (p *RedisPool) getQuestionList(eventID string, start int, end int, sort string, order string) (questionList QuestionList) {
	redisConn := p.getRedisConnection()
	defer redisConn.Close()

	/* Redisにデータが存在するか確認する。 */
	questionsKey, likeSortedKey, createdSortedKey := getQuestionsKey(eventID)

	// 3種類のKeyが存在しない場合はデータが何かしら不足しているため、データの同期を行う
	hasQuestionsKey := redisHasKey(redisConn, questionsKey)
	hasLikeSortedKey := redisHasKey(redisConn, likeSortedKey)
	hasCreatedSortedKey := redisHasKey(redisConn, createdSortedKey)

	if !hasQuestionsKey || !hasLikeSortedKey || !hasCreatedSortedKey {
		p.syncQuestion(eventID)
	}

	/* Redisからデータを取得する */
	// redisのcommand
	var redisCommand string
	if order == "asc" {
		redisCommand = "ZRANGE"
	} else if order == "desc" {
		redisCommand = "ZREVRANGE"
	}

	// sort redisのkey
	var sortedkey string
	if sort == "created_at" {
		sortedkey = createdSortedKey
	} else if sort == "like" {
		sortedkey = likeSortedKey
	}

	// debug
	logrus.Info(redisCommand, sortedkey, start-1, end-1)

	// API実行時に指定されたSortをRedisで実行
	var uuidSlice []string
	uuidSlice, err := redis.Strings(redisConn.Do(redisCommand, sortedkey, start-1, end-1))
	if err != nil {
		logrus.Error(err)
	}

	//for _, u := range uuidSlice {
		//logrus.Info(u)
	//}

	// RedisのDo関数は、Interface型のSliceしか受け付けないため、makeで生成 (String型のSliceはコンパイルエラー)
	// Example) HMGET questions_jks1812 questionID questionID questionID questionID ...
	var list = make([]interface{}, 0, 20)
	list = append(list, questionsKey)
	for _, str := range uuidSlice {
		list = append(list, str)
	}

	var bytesSlice [][]byte
	bytesSlice, _ = redis.ByteSlices(redisConn.Do("HMGET", list...))

	var questions []Question
	for _, bytes := range bytesSlice {
		q := new(Question)
		json.Unmarshal(bytes, q)
		questions = append(questions, *q)
	}

	// DB or Redis から取得したデータのtimezoneをAsia/Tokyoと指定
	locationTokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		logrus.Fatal(err)
	}

	for i := range questions {
		questions[i].CreatedAt = questions[i].CreatedAt.In(locationTokyo)
		questions[i].UpdatedAt = questions[i].UpdatedAt.In(locationTokyo)
	}

	questionList.Data = questions
	questionList.Object = "list"
	questionList.Type = "question"
	return questionList
}

func (p *RedisPool) syncQuestion(eventID string) {
	redisConnection := p.getRedisConnection()
	defer redisConnection.Close()

	// DBからデータを取得
	//var m sql.DBMap
	//dbmap, err := sql.InitMySQLDB()
	//err := m.InitMySQLDB()
	var m *gorp.DbMap
	db, err := sql.InitMySQL()
	if err != nil {
		logrus.Error(err)
		return
	}

	m = sql.MappingDBandTable(db)

	//dbmap.AddTableWithName(Question{}, "questions")
	m.AddTableWithName(Question{}, "questions")
	//defer dbmap.Db.Close()
	defer m.Db.Close()

	if err != nil {
		causeErr := errors.Cause(err)
		fmt.Printf("%+v", causeErr)
		return
	}

	var questions []Question
	//_, err = dbmap.Select(&questions, "SELECT * FROM questions WHERE event_id = '"+eventID+"'")
	_, err = m.Select(&questions, "SELECT * FROM questions WHERE event_id = '"+eventID+"'")

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
		//fmt.Println(questionsKey, " ", question.ID, " ", string(serializedJSON))
		redisConnection.Do("HSET", questionsKey, question.ID, serializedJSON)

		//SortedSet(Like)
		redisConnection.Do("ZADD", likeSortedKey, question.Like, question.ID)

		//SortedSet(CreatedAt)
		redisConnection.Do("ZADD", createdSortedKey, question.CreatedAt.Unix(), question.ID)
	}
}

func getQuestionsKey(eventID string) (questionsKey string, likeSortedKey string, createdSortedKey string) {
	questionsKey = "questions_" + eventID
	likeSortedKey = questionsKey + "_like"
	createdSortedKey = questionsKey + "_created"

	return questionsKey, likeSortedKey, createdSortedKey
}

// redisHasKey
func redisHasKey(conn redis.Conn, key string) bool {
	hasInt, err := redis.Int(conn.Do("EXISTS", key))
	if err != nil {
		logrus.Error(err)
	}

	var hasKey bool
	if hasInt == 1 {
		hasKey = true
	} else {
		hasKey = false
	}

	return hasKey
}
