package redis

import (
	"context"
	"github.com/ziyoufei11/go-tool/support/logger"
	"time"

	"github.com/spf13/cast"
)

type lockApi struct {
	lockKey string
	uniKey  string
	waitKey string
}

func LockInit(lockKey string, uniKey string) *lockApi {
	//兼容分布式
	lockKey = "{" + lockKey + "}"
	return &lockApi{
		lockKey: lockKey,
		uniKey:  uniKey,
		waitKey: lockKey + "_WAIT",
	}
}

func (l *lockApi) Lock(lockTime int, timeout float64) bool {
	//锁key时间,不允许死锁
	if lockTime <= 0 {
		lockTime = 3
	}
	for {
		resTime, next, flag := calLock(l.lockKey, l.uniKey, lockTime, timeout)
		if !next {
			return flag
		}
		timeout = resTime
	}
}

// 剩余时间 是否下一次 是否抢成功
// timeout <0 永久抢锁 =0 抢一次 >0 抢指定时间.超时退出
func calLock(lockKey string, uniKey string, lockTime int, timeout float64) (float64, bool, bool) {
	//加锁
	result, err := Client().SetNX(context.TODO(), lockKey, uniKey, time.Duration(lockTime)*time.Second).Result()
	if err == nil && result {
		//成功,返回
		return 0, false, true
	}
	//超时退出
	if timeout == 0 {
		return 0, false, false
	}
	var wait = timeout
	//获取锁剩余时间
	duration, err := Client().TTL(context.TODO(), lockKey).Result()
	if err == nil && duration.Milliseconds() > 0 {
		ttime := cast.ToFloat64(duration.Milliseconds() / 1000)
		if ttime < timeout {
			wait = ttime
		}
	}
	popNow := time.Now().Unix()
	_, _ = Client().BLPop(context.TODO(), time.Duration(wait)*time.Second, lockKey+"_WAIT").Result()
	//不超时
	if timeout < 0 {
		return timeout, true, false
	}
	timeout = timeout - cast.ToFloat64(time.Now().Unix()-popNow)
	//尝试时间超时.结束
	if timeout < 0 {
		timeout = 0
	}
	//再次调用
	return timeout, true, false
}

func (l *lockApi) Unlock() bool {
	eval := `if redis.call('get', KEYS[1]) == ARGV[1] then
				if redis.call('del', KEYS[1]) then
					if redis.call('lLen', KEYS[2]) == 0 then
						redis.call('lpush', KEYS[2], ARGV[1]);
					end
					redis.call('expire', KEYS[2], 10);
					return 1;
				end
			end
			return 0;`
	_, err := Client().Eval(context.TODO(), eval, []string{l.lockKey, l.waitKey}, l.uniKey).Result()
	if err != nil {
		logger.SugarLog.Errorf("redis锁释放失败:" + err.Error())
		return false
	}
	return true
}
