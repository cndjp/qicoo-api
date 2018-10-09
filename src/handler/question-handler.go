// QuestionHandlerの共通部分定義
// 主にRedisに関するinterfaceやstructを定義

package handler

import (
	"time"

	"github.com/cndjp/qicoo-api/src/pool"
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

// MuxVars RequestURLを格納するstruct
type MuxVars struct {
	EventID string
	Start   int
	End     int
	Sort    string
	Order   string
}

// RedisClientInterface RedisConnectionを扱うinterface
// testコードのために使用
type RedisClientInterface interface {
	GetRedisConnection() (conn redis.Conn)
	selectRedisCommand() (redisCommand string)
	selectRedisSortedKey() (sortedkey string)
	GetQuestionList() (questionList QuestionList)
	SetQuestion(question Question) error
	DeleteQuestion(question Question) error
	getQuestionsKey()
	checkRedisKey()
	syncQuestion(eventID string)
}

// RedisClient interfaceを実装するstruct
type RedisClient struct {
	Vars             MuxVars
	RedisConn        redis.Conn
	QuestionsKey     string
	LikeSortedKey    string
	CreatedSortedKey string
}

// GetInterfaceRedisConnection RedisClientからConnectionを取得
func GetInterfaceRedisConnection(rci RedisClientInterface) (conn redis.Conn) {
	return rci.GetRedisConnection()
}

// GetRedisConnection RedisのPoolから、Connectionを取得
func (rc *RedisClient) GetRedisConnection() (conn redis.Conn) {
	return pool.RedisPool.Get()
}

func (rc *RedisClient) getQuestionsKey() {
	rc.QuestionsKey = "questions_" + rc.Vars.EventID
	rc.LikeSortedKey = rc.QuestionsKey + "_like"
	rc.CreatedSortedKey = rc.QuestionsKey + "_created"

	return
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
