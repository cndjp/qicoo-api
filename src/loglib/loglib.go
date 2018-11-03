package loglib

import (
	"go.uber.org/zap"
)

var sugar *zap.SugaredLogger

// GetSugar SingleパターンでSugarオブジェクトを取得する
func GetSugar() *zap.SugaredLogger {
	if sugar == nil {
		logger, _ := zap.NewDevelopment()
		sugar = logger.Sugar()
	}
	return sugar
}
