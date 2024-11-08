package main

import (
	"encoding/json"
	"errors"

	"com.derso.aprendendo/conexoes"
	hotel "com.derso.aprendendo/hotel/negocio"
	"com.derso.aprendendo/sagas"
	"gorm.io/gorm"
)

const (
	CONFIRMACAO = 1
	TIMEOUT     = 2
)

type Mensagem struct {
	IdVaga string `json:"idVaga"`
	Acao   byte   `json:"acao"`
}

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

	msgHotel := Mensagem{}
	err := json.Unmarshal([]byte(mensagem.Dados), &msgHotel)

	if err != nil {
		return err
	}

	// TODO Tratamento do erro na sessão
	// Ao final do processo de reversão, é preciso reverter o estado para PRE_RESERVA
	return gormPostgres.Transaction(func(tx *gorm.DB) error {
		if mensagem.Tipo == sagas.EXECUTE {
			switch msgHotel.Acao {
			case CONFIRMACAO:
				return hotel.Reservar(tx, msgHotel.IdVaga)
			case TIMEOUT:
				return hotel.Liberar(tx, msgHotel.IdVaga)
			default:
				return errors.New("ação não definida")
			}
		} else {
			// Não está previsto reverter timeout
			return hotel.ReverterReserva(tx, msgHotel.IdVaga)
		}
	})
}
