package handler_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
)

var testRedisConn *redigomock.Conn

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	conn := redigomock.NewConn()
	defer func() {
		conn.Clear()
		err := conn.Close()
		if err != nil {
			log.Fatal("runTests: failed launch redis server:", err)
		}
	}()

	testRedisConn = conn

	return m.Run()
}

func flushRedis() {
	err := testRedisConn.Flush()
	if err != nil {
		fmt.Println(err)
		return
	}
}

type Person struct {
	Name string `redis:"name"`
	Age  int    `redis:"age"`
}

func RetrievePerson(conn redis.Conn, id string) (Person, error) {
	var person Person

	values, err := redis.Values(conn.Do("HGETALL", fmt.Sprintf("person:%s", id)))
	if err != nil {
		return person, err
	}

	err = redis.ScanStruct(values, &person)
	return person, err
}

func TestGetQuestionList(t *testing.T) {
	defer flushRedis()
	//多分GetQuestionListから*redis.Connを取り出して組み直さないとテストはできない。
}
