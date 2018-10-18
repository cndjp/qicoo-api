package handler

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// ReadinessHandler ReadinessProbe用function
func ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	var rci RedisConnectionInterface
var rci RedisConnectionInterface = new(RedisManager)

	var dmi MySQLDbmapInterface
	dmi = new(MySQLManager)

	err := ReadinessFunc(rci, dmi)
	if err != nil {
		logrus.Error(err)
		return
	}

	w.Write([]byte("OK"))
}

// ReadinessFunc test
func ReadinessFunc(rci RedisConnectionInterface, dmi MySQLDbmapInterface) error {
	// DBに接続可能か確認
	dbmap := dmi.GetMySQLdbmap()
	defer dbmap.Db.Close()

	_, err := dbmap.Query("SELECT 1;")
	if err != nil {
		logrus.Error(err)
		return err
	}

	//Redisに接続可能か確認
	redisConn := rci.GetRedisConnection()
	defer redisConn.Close()

	_, err = redisConn.Do("SETEX", "readiness", "10800", "readinessmsg")
	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}
