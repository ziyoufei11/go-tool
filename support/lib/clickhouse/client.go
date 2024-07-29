package clickhouse

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"sync"
	"time"
)

var isConnected bool
var CkConnect driver.Conn
var once sync.Once

func InitClickHouse() {
	once.Do(func() {
		connectToCK()
		go heartbeat()
	})
	return
}

func connectToCK() (err error) {
	ctx := context.Background()
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: viper.GetStringSlice("clickhouse.hosts"),
		Auth: clickhouse.Auth{
			Database: viper.GetString("clickhouse.database"),
			Username: viper.GetString("clickhouse.username"),
			Password: viper.GetString("clickhouse.password"),
		},
		Debugf: func(format string, v ...interface{}) {
			fmt.Printf(format, v)
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	if err = conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		fmt.Println(err)
		return
	}
	CkConnect = conn
	return
}
func heartbeat() {
	for {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			if !isConnected {
				err := CkConnect.Ping(context.TODO())
				if err != nil {
					fmt.Printf("[CK] Error: %v \n", err)
					err = connectToCK()
					if err == nil {
						fmt.Println("[CK] Reconnected")
					} else {
						fmt.Println("[CK] Reconnected Error", zap.Error(err))
					}
				}
			}
		}
	}
}
