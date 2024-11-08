package main

import (
	"com.derso.aprendendo/conexoes"
	hotel "com.derso.aprendendo/hotel/negocio"
	"com.derso.aprendendo/sagas"
	"gorm.io/gorm"
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

	gormPostgres.AutoMigrate(&hotel.Vaga{})

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
			return executar(gormPostgres, mensagem)
		},
		estaFila,
		nil,
		&proximaFila,
	)
}

func executar(gormPostgres *gorm.DB, mensagem sagas.Mensagem) error {
	// TODO a obtenção do lock da sessão e a mudança do estado devem ser feitas
	// antes de enviar a mensagem para a fila

	/*
		bloqueio, err := redisLock.Bloquear(mensagemHotel.IdSessao)

		if err != nil {
			return err
		}

		defer redisLock.Desbloquear(bloqueio)
	*/

	// TODO É preciso diferenciar o processo sendo iniciado: confirmação ou timeout
	// Usar json.Unmarshal
	idVaga := mensagem.Dados

	return gormPostgres.Transaction(func(tx *gorm.DB) error {
		if mensagem.Tipo == sagas.EXECUTE {
			return hotel.Confirmar(tx, idVaga)
		} else {
			return hotel.Cancelar(tx, idVaga)
		}
	})
}
