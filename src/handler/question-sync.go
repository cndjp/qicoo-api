package handler

import (
	"encoding/json"
	"time"

	"github.com/go-gorp/gorp"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

// これ並列化できる（チャンネル込みで）
func checkRedisKey(conn redis.Conn, rks RedisKeys) bool {
	// 3種類のKeyが存在しない場合はデータが何かしら不足しているため、データの同期を行う
	if !redisHasKey(conn, rks.QuestionKey) || !redisHasKey(conn, rks.LikeSortedKey) || !redisHasKey(conn, rks.CreatedSortedKey) {
		return false
	}

	return true
}

func syncQuestion(conn redis.Conn, m *gorp.DbMap, eventID string, rks RedisKeys) {
	var questions []Question
	_, err := m.Select(&questions, "SELECT * FROM questions WHERE event_id = '"+eventID+"'")
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

	//DBのデータをRedisに同期する。
	for _, question := range questions {
		//HashMap SerializedされたJSONデータを格納
		serializedJSON, err := json.Marshal(question)
		if err != nil {
			logrus.Error(err)
			return
		}

		if _, err := conn.Do("HSET", rks.QuestionKey, question.ID, serializedJSON); err != nil {
			logrus.Error(err)
			return
		}

		//SortedSet(Like)
		if _, err := conn.Do("ZADD", rks.LikeSortedKey, question.Like, question.ID); err != nil {
			logrus.Error(err)
			return
		}

		//SortedSet(CreatedAt)
		if _, err := conn.Do("ZADD", rks.CreatedSortedKey, question.CreatedAt.Unix(), question.ID); err != nil {
			logrus.Error(err)
			return
		}
	}
}
