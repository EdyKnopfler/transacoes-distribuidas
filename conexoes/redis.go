package conexoes

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis"
	goredisRedsync "github.com/go-redsync/redsync/v4/redis/goredis/v9"
	goredislib "github.com/redis/go-redis/v9"
)

type RedisConnection struct {
	client *goredislib.Client
	pool   redis.Pool
	sync   *redsync.Redsync
}

const LOCK_DATABASE = 1
const SESSION_DATABASE = 2

func ConectarRedis(database int) *RedisConnection {
	client := goredislib.NewClient(&goredislib.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       database,
	})

	pool := goredisRedsync.NewPool(client)
	sync := redsync.New(pool)

	return &RedisConnection{
		client: client,
		pool:   pool,
		sync:   sync,
	}
}

func (conn *RedisConnection) Fechar() {
	conn.client.Close()
}

func (conn *RedisConnection) Bloquear(chave string) (*redsync.Mutex, error) {
	mutex := conn.sync.NewMutex(chave)

	if err := mutex.Lock(); err != nil {
		return nil, err
	}

	return mutex, nil
}

func (conn *RedisConnection) Desbloquear(mutex *redsync.Mutex) (bool, error) {
	status, err := mutex.Unlock()
	return status, err
}

func (conn *RedisConnection) Setar(key string, value string) error {
	return conn.client.Set(context.Background(), key, value, 0).Err()
}

func (conn *RedisConnection) SetarObjeto(key string, keyValues ...any) error {
	return conn.client.HSet(context.Background(), key, keyValues...).Err()
}

func (conn *RedisConnection) Obter(key string) (string, error) {
	val, err := conn.client.Get(context.Background(), key).Result()

	if err == goredislib.Nil {
		err = nil
	}

	return val, err
}

func (conn *RedisConnection) ObterObjeto(key string, ponteiroObj interface{}) error {
	query := conn.client.HGetAll(context.Background(), key)

	if _, err := query.Result(); err != nil {
		return err
	}

	if err := query.Scan(ponteiroObj); err != nil {
		return err
	}

	return nil
}

func (conn *RedisConnection) Iterar(tamanhoBatch int64, fnExec func(key string) error, fnErro func(key string, err error)) error {
	var cursor uint64

	for {
		batchKeys, cursor, err := conn.client.Scan(context.Background(), cursor, "*", tamanhoBatch).Result()

		if err != nil {
			return err
		}

		for _, key := range batchKeys {
			err = fnExec(key)

			if err != nil {
				fnErro(key, err)
			}
		}

		if cursor == 0 {
			break
		}
	}

	return nil
}

func (conn *RedisConnection) Deletar(key string) error {
	return conn.client.Del(context.Background(), key).Err()
}
