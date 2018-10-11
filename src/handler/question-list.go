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

	v := QuestionListMuxVars{
		EventID: vars["event_id"],
		Start:   start,
		End:     end,
		Sort:    vars["sort"],
		Order:   vars["order"],
	}

	var rci RedisConnectionInterface
	rci = new(RedisManager)

	// QuestionListを取得
	var questionList QuestionList
	questionList = QuestionListFunc(rci, v)

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
func QuestionListFunc(rci RedisConnectionInterface, v QuestionListMuxVars) (questionList QuestionList) {
	// RedisのConnection生成
	redisConn := rci.GetRedisConnection()
	defer redisConn.Close()

	// RedisKey取得
	var rks RedisKeys
	rks = GetRedisKeys(v.EventID)

	// 多分並列処理できるやつ
	/* Redisにデータが存在するか確認する。 */
	yes := checkRedisKey(redisConn, rks)
	if !yes {
		dbmap := InitMySQLQuestion()
		syncQuestion(redisConn, dbmap, v.EventID, rks)
	}

	questionList = GetQuestionList(redisConn, v, rks)
	return questionList
}

func selectRedisCommand(order string) (redisCommand string) {
	switch order {
	case "asc":
		return "ZRANGE"
	case "desc":
		return "ZREVRANGE"
	default:
		return "ZRANGE"
	}
}

func selectRedisSortedKey(sort string, rks RedisKeys) (sortedkey string) {
	switch sort {
	case "created_at":
		return rks.CreatedSortedKey
	case "like":
		return rks.LikeSortedKey
	default:
		return rks.LikeSortedKey
	}
}

// GetQuestionList RedisとDBからデータを取得する
func GetQuestionList(conn redis.Conn, v QuestionListMuxVars, rks RedisKeys) (questionList QuestionList) {
	// API実行時に指定されたSortをRedisで実行
	uuidSlice, err := redis.Strings(conn.Do(selectRedisCommand(v.Order), selectRedisSortedKey(v.Sort, rks), v.Start-1, v.End-1))
	println("GetQuestionList:", selectRedisCommand(v.Order), selectRedisSortedKey(v.Sort, rks), v.Start-1, v.End-1)
	if err != nil {
		logrus.Error(err)
		return
	}

	// RedisのDo関数は、Interface型のSliceしか受け付けないため、makeで生成 (String型のSliceはコンパイルエラー)
	// Example) HMGET questions_jks1812 questionID questionID questionID questionID ...
	var list = make([]interface{}, 0, 20)
	list = append(list, rks.QuestionKey)
	for _, str := range uuidSlice {
		list = append(list, str)
	}

	bytesSlice, err := redis.ByteSlices(conn.Do("HMGET", list...))
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
