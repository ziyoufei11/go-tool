package redis

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"

	"go_tool/support/logger"
	"go_tool/support/runtime"
)

type Config struct {
	Addr         string        `yaml:"addr" json:"addr"`
	Username     string        `yaml:"username" json:"username"`
	Pwd          string        `yaml:"pwd" json:"pwd"`
	Db           int           `yaml:"db" json:"db"`
	MinIdleConns int           `yaml:"minIdleConns" json:"minIdleConns"`           //闲置连接的最小数量，在建立新连接速度较慢时很有用。
	IdleTimeout  time.Duration `yaml:"idleTimeout" json:"idleTimeout,omitempty"`   //客户端关闭空闲连接的时间,默认5分钟，-1关闭配置
	DialTimeout  time.Duration `yaml:"dialTimeout" json:"dialTimeout,omitempty"`   //客户端关闭空闲连接的时间,默认5分钟，-1关闭配置
	ReadTimeout  time.Duration `yaml:"readTimeout" json:"readTimeout,omitempty"`   //客户端关闭空闲连接的时间,默认5分钟，-1关闭配置
	WriteTimeout time.Duration `yaml:"writeTimeout" json:"writeTimeout,omitempty"` //客户端关闭空闲连接的时间,默认5分钟，-1关闭配置
	Env          string        `yaml:"env" json:"env"`
}

var (
	Store     *Redis
	redisOnce sync.Once
)

type CloseCmdable interface {
	redis.Cmdable
	Close() error
}

type Redis struct {
	Client        *redis.Client
	ClusterClient *redis.ClusterClient
	Config        *Config
	configPrefix  string
	IsCluster     bool
	Ctx           context.Context
	Err           error
	CloseCmdable  CloseCmdable
}

func Init() {
	redisOnce.Do(func() {
		Store = &Redis{
			Ctx:          context.Background(),
			Config:       &Config{},
			configPrefix: "redis",
		}

		err := viper.UnmarshalKey(Store.configPrefix, &Store.Config)
		if err != nil {
			Store.Err = err
		}

		isCluster := strings.Contains(Store.Config.Addr, ",")

		if isCluster {
			Store.connectToRedisCluster()
		} else {
			Store.connectToRedis()
		}

		go Store.guard()
	})

}

func Z(score float64, member interface{}) redis.Z {
	return redis.Z{
		Score:  score,
		Member: member,
	}
}

func (r *Redis) connectToRedisCluster() *redis.ClusterClient {

	r.ClusterClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        strings.Split(r.Config.Addr, ","),
		Username:     r.Config.Username,
		Password:     r.Config.Pwd,
		MinIdleConns: r.Config.MinIdleConns,
		IdleTimeout:  r.Config.IdleTimeout,
		//DialTimeout:  r.Config.DialTimeout,  // 设置连接超时
		//ReadTimeout:  r.Config.ReadTimeout,  // 设置读取超时
		//WriteTimeout: r.Config.WriteTimeout, // 设置写入超时
	})
	r.IsCluster = true

	err := Store.CanIUseRedis()
	if err != nil {
		r.Err = err
	}
	r.CloseCmdable = r.ClusterClient

	return r.ClusterClient
}

func (r *Redis) connectToRedis() *redis.Client {
	r.Client = redis.NewClient(&redis.Options{
		Addr:         r.Config.Addr,
		Username:     r.Config.Username,
		Password:     r.Config.Pwd,
		DB:           r.Config.Db,
		MinIdleConns: r.Config.MinIdleConns,
		IdleTimeout:  r.Config.IdleTimeout,
	})

	err := Store.CanIUseRedis()
	if err != nil {
		r.Err = err
	}
	r.CloseCmdable = r.Client

	return r.Client
}

func Client() CloseCmdable {
	if Store.Err != nil {
		runtime.SysStatus = runtime.SysRedisError
	} else {
		runtime.SysStatus = runtime.SysOK
	}
	return Store.CloseCmdable
}

func Key(key string) string {
	return Store.Config.Env + ":" + key
}

func (r *Redis) Use() *redis.Client {
	if r.Err != nil {
		runtime.SysStatus = runtime.SysRedisError
	} else {
		runtime.SysStatus = runtime.SysOK
	}

	return r.Client
}

func (r *Redis) UseCluster() *redis.ClusterClient {
	if r.Err != nil {
		runtime.SysStatus = runtime.SysRedisError
	} else {
		runtime.SysStatus = runtime.SysOK
	}

	return r.ClusterClient
}

func (r *Redis) guard() {
	//检查redis是否正常工作
	timer := time.NewTicker(10 * time.Second)

	for {
		<-timer.C

		err := r.CanIUseRedis()
		if err != nil {
			//记录异常信息
			logger.SugarLog.Errorf("redis状态异常: %s", err.Error())

			//重新连接到redis
			r.RetryConnectToRedis()
		}
	}
}

func (r *Redis) RetryConnectToRedis() {
	//关闭之前的连接
	var err error
	if r.IsCluster {
		err = r.ClusterClient.Close()
		if err != nil {
			logger.SugarLog.Errorf("重试时关闭redis连接失败 %s", err.Error())
		}
		r.connectToRedisCluster()
	} else {
		err = r.Client.Close()
		if err != nil {
			logger.SugarLog.Errorf("重试时关闭redis连接失败 %s", err.Error())
		}
		//重新连接到redis
		r.connectToRedis()
	}

}

func (r *Redis) CanIUseRedis() error {
	ctx, cancel := context.WithTimeout(r.Ctx, time.Second*5)
	defer cancel()
	var err error
	if r.IsCluster {
		_, err = r.ClusterClient.Ping(ctx).Result()
	} else {
		_, err = r.Client.Ping(ctx).Result()
	}

	if err != nil {
		r.Err = err
	} else {
		r.Err = nil
	}

	return err
}
