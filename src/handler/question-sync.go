package handler

import (
	"encoding/json"
	"time"

	"github.com/cndjp/qicoo-api/src/sql"
	"github.com/go-gorp/gorp"
	"github.com/sirupsen/logrus"
)

// これ並列化できる（チャンネル込みで）
func (rc *RedisClient) checkRedisKey() {
	// 3種類のKeyが存在しない場合はデータが何かしら不足しているため、データの同期を行う
	if !redisHasKey(rc.RedisConn, rc.QuestionsKey) || !redisHasKey(rc.RedisConn, rc.LikeSortedKey) || !redisHasKey(rc.RedisConn, rc.CreatedSortedKey) {
		rc.syncQuestion(rc.Vars.EventID)
	}

}

func (rc *RedisClient) syncQuestion(eventID string) {
	//	redisConnection := p.GetInterfaceRedisConnection()
	//	defer redisConnection.Close()

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

	var questions []Question
	_, err = m.Select(&questions, "SELECT * FROM questions WHERE event_id = '"+eventID+"'")
	if err != nil {
		logrus.Error(err)
		return
	}

	// DB or Redis から取得したデータのtimezoneをUTCからAsia/Tokyoと指定
	locationTokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		logrus.Error(err)
		return
	}

	for i := range questions {
		questions[i].CreatedAt = questions[i].CreatedAt.In(locationTokyo)
		questions[i].UpdatedAt = questions[i].UpdatedAt.In(locationTokyo)
	}

	//Redisで利用するKeyを取得
	rc.getQuestionsKey()

	//DBのデータをRedisに同期する。
	for _, question := range questions {
		//HashMap SerializedされたJSONデータを格納
		serializedJSON, err := json.Marshal(question)
		if err != nil {
			logrus.Error(err)
			return
		}

		if _, err := rc.RedisConn.Do("HSET", rc.QuestionsKey, question.ID, serializedJSON); err != nil {
			logrus.Error(err)
			return
		}

		//SortedSet(Like)
		if _, err := rc.RedisConn.Do("ZADD", rc.LikeSortedKey, question.Like, question.ID); err != nil {
			logrus.Error(err)
			return
		}

		//SortedSet(CreatedAt)
		if _, err := rc.RedisConn.Do("ZADD", rc.CreatedSortedKey, question.CreatedAt.Unix(), question.ID); err != nil {
			logrus.Error(err)
			return
		}
	}
}
