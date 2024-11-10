package main

import (
	"errors"

	"com.derso.aprendendo/conexoes"
	hotel "com.derso.aprendendo/hotel/negocio"
	"com.derso.aprendendo/sagas"
	"com.derso.aprendendo/sessoes"
	"gorm.io/gorm"
)

const (
	CONFIRMACAO float64 = 1
	TIMEOUT     float64 = 2
)

type Mensagem struct {
	IdVaga   string `json:"idVaga"`
	IdSessao string `json:"idSessao"`
	Acao     byte   `json:"acao"`
}

var (
	gormPostgres *gorm.DB
	redisSessoes *conexoes.RedisConnection
	redisLock    *conexoes.RedisConnection
)

func main() {
	db, err := conexoes.ConectarPostgreSQL("hotel")

	if err != nil {
		panic("Não foi possível conectar-se ao PostgreSQL.")
	}

	gormPostgres = db

	defer func() {
		sqlDB, _ := gormPostgres.DB()
		sqlDB.Close()
	}()

	gormPostgres.AutoMigrate(&hotel.Vaga{})

	redisSessoes = conexoes.ConectarRedis(conexoes.SESSION_DATABASE)
	defer redisSessoes.Fechar()

	redisLock = conexoes.ConectarRedis(conexoes.LOCK_DATABASE)
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
		func(mensagem sagas.Mensagem) (err error) {
			err = executar(mensagem)
			return
		},
		estaFila,
		nil,
		&proximaFila,
	)
}

func executar(mensagem sagas.Mensagem) error {
	// TODO a obtenção do lock da sessão e a mudança do estado devem ser feitas
	// antes de enviar a mensagem para a fila

	/*
		bloqueio, err := redisLock.Bloquear(mensagemHotel.IdSessao)

		if err != nil {
			return err
		}

		defer redisLock.Desbloquear(bloqueio)
	*/

	return gormPostgres.Transaction(func(tx *gorm.DB) error {
		if mensagem["tipo"] == sagas.EXECUTE {
			switch mensagem["acao"] {
			case CONFIRMACAO:
				// Tratamento de erro feito no módulo sagas; aqui vamos mais despreocupados :)
				return hotel.Reservar(tx, mensagem["idVaga"].(string))
			case TIMEOUT:
				return hotel.Liberar(tx, mensagem["idVaga"].(string))
			default:
				return errors.New("ação não definida")
			}
		} else {
			// Não está previsto reverter timeout
			return sessoes.ExecutarSobBloqueio(mensagem["idSessao"].(string), redisLock, func() error {
				err := hotel.ReverterReserva(tx, mensagem["idVaga"].(string))

				if err != nil {
					return err
				}

				return sessoes.MudarEstado(mensagem["idSessao"].(string), redisSessoes, sessoes.ATIVA)
			})
		}
	})
}
