package main

import (
	"fmt"

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

	filaAnterior := "hotel"
	estaFila := "passagens"
	proximaFila := "pagamentos"
	sagas.Inicializar(rabbitMQ.Channel)
	sagas.ConfigurarServicoSagas(rabbitMQ.Channel, estaFila, &filaAnterior, &proximaFila)

	sagas.IniciarConsumo(
		rabbitMQ.Channel,
		estaFila,
		func(mensagem sagas.Mensagem) error {
			if mensagem.Tipo == sagas.EXECUTE {
				fmt.Println("Executando " + string(mensagem.Dados))
			} else {
				fmt.Println("Revertendo " + string(mensagem.Dados))
			}

			return nil
		},
		estaFila,
		&filaAnterior,
		&proximaFila,
	)
}
