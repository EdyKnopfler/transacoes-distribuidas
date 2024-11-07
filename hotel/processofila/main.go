package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"com.derso.aprendendo/conexoes"
	"com.derso.aprendendo/sagas"
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

	redisSessoes := conexoes.ConectarRedis(conexoes.SESSION_DATABASE)
	defer redisSessoes.Fechar()

	redisLock := conexoes.ConectarRedis(conexoes.LOCK_DATABASE)
	defer redisLock.Fechar()

	rabbitMQ, err := conexoes.ConectarRabbitMQ()

	if err != nil {
		panic("Não foi possível conectar-se ao RabbitMq.")
	}

	defer rabbitMQ.Fechar()

	estaFila := "hotel"
	proximaFila := "passagens"
	sagas.Inicializar(rabbitMQ.Channel)
	sagas.ConfigurarServicoSagas(rabbitMQ.Channel, estaFila, nil, &proximaFila)

	sagas.IniciarConsumo(
		rabbitMQ.Channel,
		estaFila,
		func(mensagem sagas.Mensagem) error {
			if mensagem.Tipo == sagas.EXECUTE {
				fmt.Println("Executando " + string(mensagem.Dados))
				return nil
			} else {
				msg := "Erro ao reverter " + string(mensagem.Dados)
				fmt.Println(msg)
				return errors.New(msg)
			}
		},
		estaFila,
		nil,
		&proximaFila,
	)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, os.Interrupt) // os.Interrupt: Ctrl+C
	<-stop
}
