// QuestionHandlerの共通部分定義
// 主にRedisに関するinterfaceやstructを定義

package handler

import (
	"time"

	"github.com/cndjp/qicoo-api/src/mysqlib"
	"github.com/cndjp/qicoo-api/src/pool"
	"github.com/go-gorp/gorp"
	"github.com/gomodule/redigo/redis"
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
func redisHasKey(conn redis.Conn, key string) (hasKey bool) {
	ok, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		logrus.Error(err)
		return false
	}

	return ok
}

//MySQLManager MySQLDbmapInterfaceの実装
type MySQLManager struct {
}

// GetMySQLdbmap DBのdbmapを取得
func (mm *MySQLManager) GetMySQLdbmap() *gorp.DbMap {
	dbmap, err := mysqlib.InitMySQL()

	if err != nil {
		logrus.Error(err)
		return nil
	}

	dbmap.AddTableWithName(Question{}, "questions")
	return dbmap
}
