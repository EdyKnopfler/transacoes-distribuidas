package main

import (
	"errors"

	"com.derso.aprendendo/conexoes"
	passagens "com.derso.aprendendo/passagens/negocio"
	"com.derso.aprendendo/sagas"
	"gorm.io/gorm"
)

const (
	CONFIRMACAO float64 = 1
	TIMEOUT     float64 = 2
)

type Mensagem struct {
	IdAssento string `json:"idAssento"`
	IdSessao  string `json:"idSessao"`
	Acao      byte   `json:"acao"`
}

var gormPostgres *gorm.DB

func main() {
	gormPostgres, err := conexoes.ConectarPostgreSQL("passagens")

	if err != nil {
		panic("Não foi possível conectar-se ao PostgreSQL.")
	}

	defer func() {
		sqlDB, _ := gormPostgres.DB()
		sqlDB.Close()
	}()

	gormPostgres.AutoMigrate(&passagens.Assento{})

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
			return executar(mensagem)
		},
		estaFila,
		&filaAnterior,
		&proximaFila,
	)
}

func executar(mensagem sagas.Mensagem) error {
	return gormPostgres.Transaction(func(tx *gorm.DB) error {
		if mensagem["tipo"] == sagas.EXECUTE {
			switch mensagem["acao"] {
			case CONFIRMACAO:
				return passagens.Reservar(tx, mensagem["idAssento"].(string))
			case TIMEOUT:
				return passagens.Liberar(tx, mensagem["idAssento"].(string))
			default:
				return errors.New("ação não definida")
			}
		} else {
			// Não está previsto reverter timeout
			return passagens.ReverterReserva(tx, mensagem["idAssento"].(string))
		}
	})
}
