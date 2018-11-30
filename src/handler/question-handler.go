// QuestionHandlerの共通部分定義
// 主にRedisに関するinterfaceやstructを定義

package handler

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cndjp/qicoo-api/src/loglib"
	"github.com/cndjp/qicoo-api/src/mysqlib"
	"github.com/cndjp/qicoo-api/src/pool"
	"github.com/go-gorp/gorp"
	"github.com/gomodule/redigo/redis"
)

func init() {
	var err error
	localLocation, err = time.LoadLocation(localTimeZone)
	if err != nil {
		panic(fmt.Sprintf("Timezone: %s is not supported.", localTimeZone))
	}
}

var (
	localLocation *time.Location
)

const (
	localTimeZone = "Asia/Tokyo"
)

// QuestionList Questionを複数格納するstruck
type QuestionList struct {
	Object string     `json:"object"`
	Type   string     `json:"type"`
	Total  int        `json:"total"`
	Data   []Question `json:"data"`
}

// Question Questionオブジェクトを扱うためのstruct
type Question struct {
	ID        string    `json:"id" db:"id, primarykey"`
	Object    string    `json:"object" db:"object"`
	Username  string    `json:"username" db:"username"`
	EventID   string    `json:"event_id" db:"event_id"`
	ProgramID string    `json:"program_id" db:"program_id"`
	Comment   string    `json:"comment" db:"comment"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Like      int       `json:"like" db:"like_count"`
}

// QuestionListMuxVars RequestURLを格納するstruct
type QuestionListMuxVars struct {
	EventID string
	Start   int
	End     int
	Sort    string
	Order   string
}

// QuestionCreateMuxVars RequestURLを格納するstruct
type QuestionCreateMuxVars struct {
	EventID string
}

// QuestionDeleteMuxVars RequestURLを格納するstruct
type QuestionDeleteMuxVars struct {
	EventID string
}

// QuestionLikeMuxVars RequestURLを格納するstruct
type QuestionLikeMuxVars struct {
	EventID    string
	QuestionID string
}

// RedisKeys Redis用のkeyを扱うstruct
type RedisKeys struct {
	QuestionKey      string //Hash
	LikeSortedKey    string //SortedSet like順
	CreatedSortedKey string //SortedSet 作成順
}

// RedisConnectionInterface RedisのConnectionを扱うInterface
type RedisConnectionInterface interface {
	GetRedisConnection() (conn redis.Conn)
}

// MySQLDbmapInterface MySQLのDBmapを扱うInterface
type MySQLDbmapInterface interface {
	GetMySQLdbmap() *gorp.DbMap
}

// RedisManager  RedisConnectionInterfaceの実装
type RedisManager struct {
}

// GetRedisConnection RedisConnectionの取得
func (rm *RedisManager) GetRedisConnection() (conn redis.Conn) {
	return pool.RedisPool.Get()
}

// GetRedisKeys Redisで使用するkeyを取得
func GetRedisKeys(eventID string) RedisKeys {
	var k RedisKeys
	k.QuestionKey = "questions_" + eventID
	k.LikeSortedKey = k.QuestionKey + "_like"
	k.CreatedSortedKey = k.QuestionKey + "_created"

	return k
}

// redisHasKey
func redisHasKey(conn redis.Conn, key string) (hasKey bool, err error) {
	sugar := loglib.GetSugar()
	defer sugar.Sync()

	sugar.Infof("Redis Command of redisHasKey. command='EXISTS %s'", key)
	ok, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false, err
	}

	return ok, nil
}

//MySQLManager MySQLDbmapInterfaceの実装
type MySQLManager struct {
}

// GetMySQLdbmap DBのdbmapを取得
func (mm *MySQLManager) GetMySQLdbmap() *gorp.DbMap {
	sugar := loglib.GetSugar()
	defer sugar.Sync()

	dbmap := mysqlib.GetMySQLdbmap()
	dbmap.AddTableWithName(Question{}, "questions")
	return dbmap
}

// TimeNowRoundDown 時刻を取得する。秒未満は切り捨てる
// RedisとMySQLでの時刻扱いに微妙に仕様の差異があるための対応
// Time.Now()で生成した時刻をMySQLに挿入すると、四捨五入される
// MySQLに挿入する前に時刻を確定したいため、この関数で生成する時刻を使用する
func TimeNowRoundDown() time.Time {

	// utcからjstを生成
	utc := time.Now().UTC()
	localTime := utc.In(localLocation)
	// 秒未満を切り捨てる
	return localTime.Truncate(time.Second)
}

// getQuestion RedisからQuestionを取得する
func getQuestion(conn redis.Conn, dbmap *gorp.DbMap, eventID string, questionID string, rks RedisKeys) (Question, error) {

	var q Question

	yes, err := checkRedisKey(conn, rks)
	if err != nil {
		return q, err
	}

	if !yes {
		_, err := syncQuestion(conn, dbmap, eventID, rks)
		// 同期にエラー
		if err != nil {
			return q, err
		}
	}

	sugar := loglib.GetSugar()
	defer sugar.Sync()
	//HashからQuesitonのデータを取得する
	bytesSlice, err := redis.ByteSlices(conn.Do("HMGET", rks.QuestionKey, questionID))
	sugar.Infof("Redis Command of getQuestion. command='HMGET %s %s'", rks.QuestionKey, questionID)
	if err != nil {
		return q, err
	}

	for _, bytes := range bytesSlice {
		err = json.Unmarshal(bytes, &q)
		if err != nil {
			return q, err
		}
	}

	return q, err
}
