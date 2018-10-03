package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
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

type MuxVars struct {
	EventID string
	Start   int
	End     int
	Sort    string
	Order   string
}

type RedisPool struct {
	Pool             *redis.Pool
	Vars             MuxVars
	RedisConn        redis.Conn
	QuestionsKey     string
	LikeSortedKey    string
	CreatedSortedKey string
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
}

// GetRedisConnection
func (p *RedisPool) GetRedisConnection() (conn redis.Conn) {
	return p.Pool.Get()
}

// QuestionListHandler QuestionオブジェクトをRedisから取得する。存在しない場合はDBから取得し、Redisへ格納する
func (p *RedisPool) QuestionListHandler(w http.ResponseWriter, r *http.Request) {
	// URLに含まれている event_id を取得
	vars := mux.Vars(r)
	start, err := strconv.Atoi(vars["start"])
	if err != nil {
		logrus.Error(err)
	}
	end, err := strconv.Atoi(vars["end"])
	if err != nil {
		logrus.Error(err)
	}

	p.Vars = MuxVars{
		EventID: vars["event_id"],
		Start:   start,
		End:     end,
		Sort:    vars["sort"],
		Order:   vars["order"],
	}

	p.RedisConn = p.GetRedisConnection()
	defer p.RedisConn.Close()

	/* Redisにデータが存在するか確認する。 */
	p.getQuestionsKey()

	// 多分並列処理できるやつ
	p.checkRedisKey()

	questionList := p.GetQuestionList()

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

// GetQuestionList RedisとDBからデータを取得する
func (p *RedisPool) GetQuestionList() (questionList QuestionList) {
	/* Redisからデータを取得する */
	// redisのcommand
	var redisCommand string
	if p.Vars.Order == "asc" {
		redisCommand = "ZRANGE"
	} else if p.Vars.Order == "desc" {
		redisCommand = "ZREVRANGE"
	}

	// sort redisのkey
	var sortedkey string
	if p.Vars.Sort == "created_at" {
		sortedkey = p.CreatedSortedKey
	} else if p.Vars.Sort == "like" {
		sortedkey = p.LikeSortedKey
	}

	// debug
	logrus.Info(redisCommand, sortedkey, p.Vars.Start-1, p.Vars.End-1)

	// API実行時に指定されたSortをRedisで実行
	var uuidSlice []string
	uuidSlice, err := redis.Strings(p.RedisConn.Do(redisCommand, sortedkey, p.Vars.Start-1, p.Vars.End-1))
	if err != nil {
		logrus.Error(err)
	}

	// RedisのDo関数は、Interface型のSliceしか受け付けないため、makeで生成 (String型のSliceはコンパイルエラー)
	// Example) HMGET questions_jks1812 questionID questionID questionID questionID ...
	var list = make([]interface{}, 0, 20)
	list = append(list, p.QuestionsKey)
	for _, str := range uuidSlice {
		list = append(list, str)
	}

	var bytesSlice [][]byte
	bytesSlice, _ = redis.ByteSlices(p.RedisConn.Do("HMGET", list...))

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

func (p *RedisPool) getQuestionsKey() {
	p.QuestionsKey = "questions_" + p.Vars.EventID
	p.LikeSortedKey = p.QuestionsKey + "_like"
	p.CreatedSortedKey = p.QuestionsKey + "_created"

	return
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
