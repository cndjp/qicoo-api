package handler

import (
	"net/http"

	"github.com/cndjp/qicoo-api/src/loglib"
)

// ReadinessHandler ReadinessProbe用function
func ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	sugar := loglib.GetSugar()
	defer sugar.Sync()

	sugar.Info("Requested Readiness.")

	var rci RedisConnectionInterface = new(RedisManager)

	var dmi MySQLDbmapInterface = new(MySQLManager)

	err := ReadinessFunc(rci, dmi)
	if err != nil {
		sugar.Error(err)
		w.WriteHeader(500)
		w.Write([]byte("NG"))
	}

	w.Write([]byte("OK"))
	sugar.Info("Response Readiness.")
}

// ReadinessFunc test用に切だし
func ReadinessFunc(rci RedisConnectionInterface, dmi MySQLDbmapInterface) error {
	sugar := loglib.GetSugar()
	defer sugar.Sync()

	// DBに接続可能か確認
	dbmap := dmi.GetMySQLdbmap()
	defer dbmap.Db.Close()

	_, err := dbmap.Exec("SELECT 1;")
	if err != nil {
		sugar.Error(err)
		return err
	}

	//Redisに接続可能か確認
	redisConn := rci.GetRedisConnection()
	defer redisConn.Close()

	_, err = redisConn.Do("SETEX", "readiness", "10800", "readinessmsg")
	if err != nil {
		sugar.Error(err)
		return err
	}

	return nil
}
