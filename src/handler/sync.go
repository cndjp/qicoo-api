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

func (p *RedisPool) syncQuestion(eventID string) {
	redisConnection := p.getRedisConnection()
	defer redisConnection.Close()

	// DBからデータを取得
	//var m sql.DBMap
	//dbmap, err := sql.InitMySQLDB()
	//err := m.InitMySQLDB()
	var m *gorp.DbMap
	db, err := sql.InitMySQL()
	if err != nil {
		logrus.Error(err)
		return
	}

	m = sql.MappingDBandTable(db)

	//dbmap.AddTableWithName(Question{}, "questions")
	m.AddTableWithName(Question{}, "questions")
	//defer dbmap.Db.Close()
	defer m.Db.Close()

	if err != nil {
		causeErr := errors.Cause(err)
		fmt.Printf("%+v", causeErr)
		return
	}

	var questions []Question
	//_, err = dbmap.Select(&questions, "SELECT * FROM questions WHERE event_id = '"+eventID+"'")
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
	questionsKey, likeSortedKey, createdSortedKey := getQuestionsKey(eventID)

	//DBのデータをRedisに同期する。
	for _, question := range questions {
		//HashMap SerializedされたJSONデータを格納
		serializedJSON, _ := json.Marshal(question)
		//fmt.Println(questionsKey, " ", question.ID, " ", string(serializedJSON))
		redisConnection.Do("HSET", questionsKey, question.ID, serializedJSON)

		//SortedSet(Like)
		redisConnection.Do("ZADD", likeSortedKey, question.Like, question.ID)

		//SortedSet(CreatedAt)
		redisConnection.Do("ZADD", createdSortedKey, question.CreatedAt.Unix(), question.ID)
	}
}
