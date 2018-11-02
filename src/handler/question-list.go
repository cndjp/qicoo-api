package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/cndjp/qicoo-api/src/loglib"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
)

// QuestionListHandler QuestionオブジェクトをRedisから取得する。存在しない場合はDBから取得し、Redisへ格納する
func QuestionListHandler(w http.ResponseWriter, r *http.Request) {
	sugar := loglib.GetSugar()
	defer sugar.Sync()

	// Headerの設定
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// URLに含まれている event_id を取得
	vars := mux.Vars(r)
	start, err := strconv.Atoi(vars["start"])
	if err != nil {
		sugar.Error(err)
		return
	}
	end, err := strconv.Atoi(vars["end"])
	if err != nil {
		sugar.Error(err)
		return
	}

	v := QuestionListMuxVars{
		EventID: vars["event_id"],
		Start:   start,
		End:     end,
		Sort:    vars["sort"],
		Order:   vars["order"],
	}

	sugar.Infof("Request QuestionList process. EventID:%s, Start:%d, End:%d, Sort:%s, Order:%s", v.EventID, v.Start, v.End, v.Sort, v.Order)

	var rci RedisConnectionInterface
	rci = new(RedisManager)

	var dmi MySQLDbmapInterface
	dmi = new(MySQLManager)

	// QuestionListを取得
	var questionList QuestionList
	questionList, err = QuestionListFunc(rci, dmi, v)
	if err != nil {
		sugar.Error(err)
		return
	}

	/* JSONの整形 */
	// QuestionのStructをjsonとして変換
	jsonBytes, err := json.Marshal(questionList)
	if err != nil {
		sugar.Error(err)
		return
	}

	// 整形用のバッファを作成し、整形を実行
	out := new(bytes.Buffer)
	// プリフィックスなし、スペース2つでインデント
	err = json.Indent(out, jsonBytes, "", "  ")
	if err != nil {
		sugar.Error(err)
		return
	}

	w.Write([]byte(out.String()))
	sugar.Infof("Response QuestionList process. QuestionList:%s", jsonBytes)

}

// QuestionListFunc テストコードでテストしやすいように定義
func QuestionListFunc(rci RedisConnectionInterface, dmi MySQLDbmapInterface, v QuestionListMuxVars) (questionList QuestionList, err error) {
	sugar := loglib.GetSugar()
	defer sugar.Sync()

	// RedisのConnection生成
	redisConn := rci.GetRedisConnection()
	defer redisConn.Close()

	// RedisKey取得
	var rks RedisKeys
	rks = GetRedisKeys(v.EventID)

	// 多分並列処理できるやつ
	/* Redisにデータが存在するか確認する。 */
	yes, err := checkRedisKey(redisConn, rks)
	if err != nil {
		sugar.Error(err)
		return getZeroQuestionList(), err
	}

	if !yes {
		dbmap := dmi.GetMySQLdbmap()
		defer dbmap.Db.Close()
		sc, err := syncQuestion(redisConn, dbmap, v.EventID, rks)

		// 同期にエラー
		if err != nil {
			sugar.Error(err)
			return getZeroQuestionList(), err
		}

		// 同期したデータが0の場合(Questionデータが0の場合)
		if sc == 0 {
			return getZeroQuestionList(), nil
		}
	}

	questionList, err = GetQuestionList(redisConn, v, rks)
	if err != nil {
		sugar.Error(err)
		return questionList, err
	}

	return questionList, nil
}

// getZeroQuestionList Questionが0件の時のデータを取得
func getZeroQuestionList() QuestionList {
	var questionList QuestionList
	var questions []Question

	questionList = QuestionList{
		Data:   questions,
		Object: "list",
		Type:   "question",
	}

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
func GetQuestionList(conn redis.Conn, v QuestionListMuxVars, rks RedisKeys) (questionList QuestionList, err error) {
	sugar := loglib.GetSugar()
	defer sugar.Sync()

	// API実行時に指定されたSortをRedisで実行
	uuidSlice, err := redis.Strings(conn.Do(selectRedisCommand(v.Order), selectRedisSortedKey(v.Sort, rks), v.Start-1, v.End-1))
	sugar.Infof("Redis Command of GetQuestionList. command='%s %s %d %d'", selectRedisCommand(v.Order), selectRedisSortedKey(v.Sort, rks), v.Start-1, v.End-1)
	if err != nil {
		sugar.Error(err)
		return getZeroQuestionList(), err
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
		sugar.Error(err)
		return getZeroQuestionList(), err
	}

	var questions []Question
	for _, bytes := range bytesSlice {
		q := new(Question)
		err = json.Unmarshal(bytes, q)
		if err != nil {
			sugar.Error(err)
			return getZeroQuestionList(), err
		}

		questions = append(questions, *q)
	}

	// DB or Redis から取得したデータのtimezoneをAsia/Tokyoと指定
	locationTokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		sugar.Fatal(err)
		return getZeroQuestionList(), err
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

	return questionList, nil
}
