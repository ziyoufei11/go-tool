package mysql

import (
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"sync"
)

var instance map[string]*gorm.DB
var once sync.Once
var mu sync.Mutex

func init() {
	once.Do(func() {
		instance = make(map[string]*gorm.DB)
	})
}

func GetDb(dbName string) *gorm.DB {
	return instance[dbName]
}

func ConnectDB(dbName string, gormConfigs *gorm.Config) (db *gorm.DB, err error) {
	mu.Lock()
	defer mu.Unlock()
	db, err = gorm.Open(mysql.Open(viper.GetString(dbName+".dsn")), gormConfigs)
	if err != nil {
		return
	}
	instance[dbName] = db
	return
}
