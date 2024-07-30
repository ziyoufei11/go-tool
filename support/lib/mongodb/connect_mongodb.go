package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"sync"
	"time"
)

var sy sync.RWMutex

var MongoMap = make(map[string]*MongoInfo)

type MongoInfo struct {
	Addr        string
	Port        string
	Database    string
	UserName    string
	Password    string
	mongoClient *mongo.Client
	mongoOnce   sync.Once
}

func NewMongo(addr string, port string, database string, userName string, pwd string) *MongoInfo {
	uni := addr + port
	sy.Lock()
	defer sy.Unlock()
	if MongoMap[uni] != nil {
		return MongoMap[uni]
	}
	data := &MongoInfo{
		Addr:     addr,
		Port:     port,
		Database: database,
		UserName: userName,
		Password: pwd,
	}
	MongoMap[uni] = data
	return data
}

func (m *MongoInfo) OnceConnect() (err error) {
	m.mongoOnce.Do(func() {
		err = m.connect()
	})
	return
}

func (m *MongoInfo) connect() (err error) {
	uri := "mongodb://%s:%s/%s"
	uri = fmt.Sprintf(uri, m.Addr, m.Port, m.Database)
	o := options.Client().ApplyURI(uri)
	o.Auth = &options.Credential{
		Username: m.UserName,
		Password: m.Password,
	}
	m.mongoClient, err = mongo.Connect(context.TODO(), o)
	if err != nil {
		fmt.Println(err)
		return
	}
	go m.ping()
	return
}

func (m *MongoInfo) ping() (err error) {
	for {
		time.Sleep(time.Second * 10)
		// Ping the primary
		err = m.Client().Ping(context.TODO(), readpref.Primary())
		if err != nil {
			m.connect()
			fmt.Println("mongo ping error:", err.Error())
			break
		}
	}
	return
}

func (m *MongoInfo) Client() *mongo.Client {
	return m.mongoClient
}
