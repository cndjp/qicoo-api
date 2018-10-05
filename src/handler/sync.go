package handler

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cndjp/qicoo-api/src/sql"
	"github.com/go-gorp/gorp"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// これ並列化できる（チャンネル込みで）
func (p *RedisPool) checkRedisKey() {
	// 3種類のKeyが存在しない場合はデータが何かしら不足しているため、データの同期を行う
	if !redisHasKey(p.RedisConn, p.QuestionsKey) || !redisHasKey(p.RedisConn, p.LikeSortedKey) || !redisHasKey(p.RedisConn, p.CreatedSortedKey) {
		p.syncQuestion(p.Vars.EventID)
	}

}

func (p *RedisPool) syncQuestion(eventID string) {
	redisConnection := p.GetRedisConnection()
	defer redisConnection.Close()

	// DBからデータを取得
	var m *gorp.DbMap
	db, err := sql.InitMySQL()
	if err != nil {
		logrus.Error(err)
		return
	}

	m = sql.MappingDBandTable(db)

	m.AddTableWithName(Question{}, "questions")
	defer m.Db.Close()

	if err != nil {
		causeErr := errors.Cause(err)
		fmt.Printf("%+v", causeErr)
		return
	}

	var questions []Question
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
	p.getQuestionsKey()

	//DBのデータをRedisに同期する。
	for _, question := range questions {
		//HashMap SerializedされたJSONデータを格納
		serializedJSON, _ := json.Marshal(question)
		redisConnection.Do("HSET", p.QuestionsKey, question.ID, serializedJSON)

		//SortedSet(Like)
		redisConnection.Do("ZADD", p.LikeSortedKey, question.Like, question.ID)

		//SortedSet(CreatedAt)
		redisConnection.Do("ZADD", p.CreatedSortedKey, question.CreatedAt.Unix(), question.ID)
	}
}
