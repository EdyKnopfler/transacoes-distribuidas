package main

import (
	"com.derso.aprendendo/conexoes"
)

func main() {
	gormPostgres, err := conexoes.ConectarPostgreSQL("hotel")

	if err != nil {
		panic("Não foi possível conectar-se ao PostgreSQL.")
	}

	defer func() {
		sqlDB, _ := gormPostgres.DB()
		sqlDB.Close()
	}()

	redis := conexoes.ConectarRedis(conexoes.SESSION_DATABASE)
	defer redis.Fechar()

	rabbitMQ, err := conexoes.ConectarRabbitMQ()

	if err != nil {
		panic("Não foi possível conectar-se ao RabbitMq.")
	}

	defer rabbitMQ.Fechar()
}
