package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// QuestionListHandler QuestionオブジェクトをRedisから取得する。存在しない場合はDBから取得し、Redisへ格納する
func QuestionListHandler(w http.ResponseWriter, r *http.Request) {
	// URLに含まれている event_id を取得
	vars := mux.Vars(r)
	start, err := strconv.Atoi(vars["start"])
	if err != nil {
		logrus.Error(err)
		return
	}
	end, err := strconv.Atoi(vars["end"])
	if err != nil {
		logrus.Error(err)
		return
	}

	v := MuxVars{
		EventID: vars["event_id"],
		Start:   start,
		End:     end,
		Sort:    vars["sort"],
		Order:   vars["order"],
	}

	var questionList QuestionList
	questionList = QuestionListFunc(v)

	/* JSONの整形 */
	// QuestionのStructをjsonとして変換
	jsonBytes, err := json.Marshal(questionList)
	if err != nil {
		logrus.Error(err)
		return
	}

	// 整形用のバッファを作成し、整形を実行
	out := new(bytes.Buffer)
	// プリフィックスなし、スペース2つでインデント
	json.Indent(out, jsonBytes, "", "  ")

	w.Write([]byte(out.String()))

}

// QuestionListFunc テストコードでテストしやすいように定義
func QuestionListFunc(vars MuxVars) (questionList QuestionList) {
	// RedisClientの初期化初期設定
	rc := new(RedisClient)

	rc.Vars = vars

	rc.RedisConn = rc.GetRedisConnection()
	defer rc.RedisConn.Close()

	// 多分並列処理できるやつ
	/* Redisにデータが存在するか確認する。 */
	yes := rc.checkRedisKey()
	if !yes {
		dbmap := InitMySQLQuestion()
		rc.syncQuestion(dbmap, rc.Vars.EventID)
	}

	questionList = rc.GetQuestionList()
	return questionList
}

func (rc *RedisClient) selectRedisCommand() (redisCommand string) {
	switch rc.Vars.Order {
	case "asc":
		return "ZRANGE"
	case "desc":
		return "ZREVRANGE"
	default:
		return "ZRANGE"
	}
}

func (rc *RedisClient) selectRedisSortedKey() (sortedkey string) {
	switch rc.Vars.Sort {
	case "created_at":
		return rc.CreatedSortedKey
	case "like":
		return rc.LikeSortedKey
	default:
		return rc.LikeSortedKey
	}
}

// GetQuestionList RedisとDBからデータを取得する
func (rc *RedisClient) GetQuestionList() (questionList QuestionList) {
	rc.getQuestionsKey()

	// API実行時に指定されたSortをRedisで実行
	uuidSlice, err := redis.Strings(rc.RedisConn.Do(rc.selectRedisCommand(), rc.selectRedisSortedKey(), rc.Vars.Start-1, rc.Vars.End-1))
	println("GetQuestionList:", rc.selectRedisCommand(), rc.selectRedisSortedKey(), rc.Vars.Start-1, rc.Vars.End-1)
	if err != nil {
		logrus.Error(err)
		return
	}

	// RedisのDo関数は、Interface型のSliceしか受け付けないため、makeで生成 (String型のSliceはコンパイルエラー)
	// Example) HMGET questions_jks1812 questionID questionID questionID questionID ...
	var list = make([]interface{}, 0, 20)
	list = append(list, rc.QuestionsKey)
	for _, str := range uuidSlice {
		list = append(list, str)
	}

	bytesSlice, err := redis.ByteSlices(rc.RedisConn.Do("HMGET", list...))
	if err != nil {
		logrus.Error(err)
		return
	}

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

	questionList = QuestionList{
		Data:   questions,
		Object: "list",
		Type:   "question",
	}

	return questionList
}
