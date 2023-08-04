package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	cf "github.com/pplorins/go-mitmproxy/config"
	"github.com/redis/go-redis/v9"
	"gitlab.com/pplorins/wechat-official-accounts/chatgpt/shared"
	"regexp"
	"sync"
	"time"
)

const (
	GPT_ANSWER_MSG_KEY = "gpt-answer-%s"

	CONN_PATTERN          = "^(?P<host>.*):(?P<port>\\d+),password=(?P<pwd>.*)$"
	OPTIMISTIC_LOCK_RETRY = 100

	REDIS_TIMEOUT_SEC = 2
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
			Addr:         fmt.Sprintf("%s:%s", host, port),
			Password:     pwd,
			DialTimeout:  time.Duration(REDIS_TIMEOUT_SEC) * time.Second,
			ReadTimeout:  time.Duration(REDIS_TIMEOUT_SEC) * time.Second,
			WriteTimeout: time.Duration(REDIS_TIMEOUT_SEC) * time.Second,
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

func (r *RedisClient) txnLoop(ctx context.Context, txf func(tx *redis.Tx) error, watchKey ...string) error {
	// Retry if the key has been changed.
	for i := 0; i < OPTIMISTIC_LOCK_RETRY; i++ {
		err := r.rdb.Watch(ctx, txf, watchKey...)
		if err == nil {
			// Success.
			return nil
		}
		if err == redis.TxFailedErr {
			// Optimistic lock lost. Retry.
			continue
		}
		// Return any other error.
		return err
	}

	return errors.Errorf("increment reached maximum number of retries:%d", OPTIMISTIC_LOCK_RETRY)

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

func (r *RedisClient) WriteMidJourneyRequestHttpContext(ctx context.Context,
	taskID string,
	bc *shared.MidJourneyBaseHttpRequestContext,
	ir *shared.ImagineRequestRedis,
	ls *shared.MidJourneyLastTaskRedis,
) error {

	bk := shared.MJ_BASE_REQ_CTX_KEY
	mk := fmt.Sprintf(shared.MJ_IMAGINE_REQ_CTX_KEY, taskID)
	ck := shared.MJ_LAST_TASK_KEY

	bm, e := shared.Struct2HashFields(ctx, bc, false)
	if e != nil {
		return errors.Errorf("convert struct ChannelContext to map failed:%+v", e)
	}
	ib, e := json.Marshal(ir)
	if e != nil {
		return errors.Errorf("marshal ir failed:%+v", e)
	}

	pks := make([]string, 0, 10)
	txf := func(tx *redis.Tx) error {
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			e = pipe.HSet(ctx, bk, bm).Err()
			if e != nil {
				return errors.Errorf("hset bk failed")
			}
			e = pipe.Set(ctx, mk, ib, shared.Duration24Hour).Err()
			if e != nil {
				return errors.Errorf("hset mk failed")
			}
			if ls != nil {
				lsm, e := shared.Struct2HashFields(ctx, ls, false)
				if e != nil {
					return errors.Errorf("convert struct MsgContext to map failed:%+v", e)
				}
				e = pipe.HSet(ctx, ck, lsm).Err()
				if e != nil {
					return errors.Errorf("hset ck failed")
				}
			}
			return nil
		})
		return err
	}

	watchKeys := append(pks, bk, mk)
	return r.txnLoop(ctx, txf, watchKeys...)
}
