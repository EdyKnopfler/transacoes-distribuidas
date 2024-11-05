package conexoes

import (
	"fmt"
	"os"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	goredislib "github.com/redis/go-redis/v9"
)

type RedisConnection struct {
	client *goredislib.Client
	pool   redis.Pool
	sync   *redsync.Redsync
}

func ConectarRedis() *RedisConnection {
	client := goredislib.NewClient(&goredislib.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
	})

	pool := goredis.NewPool(client)
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

func (conn *RedisConnection) Bloquear(chave string) (error, *redsync.Mutex) {
	mutex := conn.sync.NewMutex(chave)

	if err := mutex.Lock(); err != nil {
		return err, nil
	}

	return nil, mutex
}

func (conn *RedisConnection) Desbloquear(mutex *redsync.Mutex) (bool, error) {
	status, err := mutex.Unlock()
	return status, err
}
