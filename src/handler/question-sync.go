package handler

import (
	"encoding/json"
	"time"

	"github.com/go-gorp/gorp"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

// これ並列化できる（チャンネル込みで）
func checkRedisKey(conn redis.Conn, rks RedisKeys) (bool, error) {
	// 3種類のKeyが存在しない場合はデータが何かしら不足しているため、データの同期を行う
	hasQuestion, err := redisHasKey(conn, rks.QuestionKey)
	if err != nil {
		logrus.Error(err)
		return false, err
	}

	hasLikeSorted, err := redisHasKey(conn, rks.QuestionKey)
	if err != nil {
		logrus.Error(err)
		return false, err
	}

	hasCreatedSorted, err := redisHasKey(conn, rks.QuestionKey)
	if err != nil {
		logrus.Error(err)
		return false, err
	}

	if !hasQuestion || !hasLikeSorted || !hasCreatedSorted {
		return false, nil
	}

	return true, nil
}

// syncQuestion Redisにデータが存在しない場合、MySQLと同期を行う
// return: 同期した件数(errorの場合,データが存在しない場合は0)、error
func syncQuestion(conn redis.Conn, m *gorp.DbMap, eventID string, rks RedisKeys) (int, error) {
	var questions []Question
	_, err := m.Select(&questions, "SELECT * FROM questions WHERE event_id = '"+eventID+"'")
	if err != nil {
		logrus.Error(err)
		return 0, err
	}

	// DBに何もデータが存在しない場合は、syncしない
	if len(questions) == 0 {
		return 0, nil
	}

	// DB or Redis から取得したデータのtimezoneをUTCからAsia/Tokyoと指定
	locationTokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		logrus.Error(err)
		return 0, err
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
			return 0, err
		}

		if _, err := conn.Do("HSET", rks.QuestionKey, question.ID, serializedJSON); err != nil {
			logrus.Error(err)
			return 0, err
		}

		//SortedSet(Like)
		if _, err := conn.Do("ZADD", rks.LikeSortedKey, question.Like, question.ID); err != nil {
			logrus.Error(err)
			return 0, err
		}

		//SortedSet(CreatedAt)
		if _, err := conn.Do("ZADD", rks.CreatedSortedKey, question.CreatedAt.Unix(), question.ID); err != nil {
			logrus.Error(err)
			return 0, err
		}
	}

	return len(questions), nil
}
