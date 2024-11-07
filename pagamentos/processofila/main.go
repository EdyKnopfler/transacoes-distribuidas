package main

import (
	"errors"
	"fmt"

	"com.derso.aprendendo/conexoes"
	"com.derso.aprendendo/sagas"
)

func main() {
	gormPostgres, err := conexoes.ConectarPostgreSQL("pagamentos")

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

	estaFila := "pagamentos"
	filaAnterior := "passagens"
	sagas.Inicializar(rabbitMQ.Channel)
	sagas.ConfigurarServicoSagas(rabbitMQ.Channel, estaFila, &filaAnterior, nil)

	sagas.IniciarConsumo(
		rabbitMQ.Channel,
		estaFila,
		func(mensagem sagas.Mensagem) error {
			fmt.Println("Simulando erro no pagamento")
			return errors.New("deu pau")
		},
		estaFila,
		&filaAnterior,
		nil,
	)
}
