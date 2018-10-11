package handler_test

import (
	_ "encoding/json"
	_ "log"
	_ "os"
	"testing"
	_ "time"

	_ "github.com/cndjp/qicoo-api/src/handler"
	_ "github.com/gomodule/redigo/redis"
	_ "github.com/rafaeljusto/redigomock"
)

func TestCreateQuestionRedisInTheTravis(t *testing.T) {
	//	var mockPool = newMockPool()
	//	defer func() {
	//		mockPool.RedisConn.Close()
	//
	//		// 一律でflushallはやりすぎか？
	//		flushallRedis(mockPool.RedisConn)
	//	}()
	//
	//	err := handler.CreateQuestionRedis(mockPool, mockQuestion)
	//
	//	if err != nil {
	//		t.Fatal("create question error:", err)
	//	}
}

func TestCreateQuestionRedisInTheLocal(t *testing.T) {
	//	var mockPool = newMockPool()
	//	defer func() {
	//		mockPool.RedisConn.Close()
	//
	//		// 一律でflushallはやりすぎか？
	//		internalTestRedisConn.Command("FLUSHALL").Expect("OK")
	//		flushallRedis(mockPool.RedisConn)
	//	}()
	//
	//	// redigomockのテストデータを登録。
	//	internalTestRedisConn.GenericCommand("EXISTS").Expect([]byte("true"))
	//	internalTestRedisConn.GenericCommand("HSET").Expect("OK!")
	//	internalTestRedisConn.GenericCommand("ZADD").Expect("OK!")
	//
	//	err := handler.CreateQuestionRedis(mockPool, mockQuestion)
	//
	//	if err != nil {
	//		t.Fatal("create question error:", err)
	//	}
}
