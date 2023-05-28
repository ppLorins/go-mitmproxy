package redis

import (
	"context"
	"fmt"
	cf "github.com/lqqyt2423/go-mitmproxy/config"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"regexp"
	"sync"
)

const (
	GPT_ANSWER_MSG_KEY = "gpt-answer-%s"

	CONN_PATTERN = "^(?P<host>.*):(?P<port>\\d+),password=(?P<pwd>.*)$"
)

var once sync.Once

var rdb *redis.Client

func InitializeRedis() {
	once.Do(func() {
		gc := cf.GetGlobalConfig()
		rc := gc.RedisConnString

		re := regexp.MustCompile(CONN_PATTERN)
		ma := re.FindStringSubmatch(rc)
		if len(ma) != 4 {
			msg := fmt.Sprintf("redis conn string RE match failed:%s", rc)
			panic(msg)
		}
		host := ma[1]
		port := ma[2]
		pwd := ma[3]

		rdb = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", host, port),
			Password: pwd,
		})
	})
}

func NewRedisClient() *RedisClient {
	return &RedisClient{
		rdb: rdb,
	}
}

type RedisClient struct {
	rdb *redis.Client
}

func (r *RedisClient) NotifyAnswer(ctx context.Context, sessionID, answer string) error {
	k := fmt.Sprintf(GPT_ANSWER_MSG_KEY, sessionID)
	c := r.rdb.RPush(ctx, k, answer)
	e := c.Err()
	if e != nil {
		return errors.Errorf("redis rpush failed:%+v", e)
	}

	return nil
}
