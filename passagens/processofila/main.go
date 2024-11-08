package main

import (
	"encoding/json"
	"errors"

	"com.derso.aprendendo/conexoes"
	passagens "com.derso.aprendendo/passagens/negocio"
	"com.derso.aprendendo/sagas"
	"gorm.io/gorm"
)

const (
	CONFIRMACAO = 1
	TIMEOUT     = 2
)

type Mensagem struct {
	IdAssento string `json:"idAssento"`
	Acao      byte   `json:"acao"`
}

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
			return executar(gormPostgres, mensagem)
		},
		estaFila,
		&filaAnterior,
		&proximaFila,
	)
}

func executar(gormPostgres *gorm.DB, mensagem sagas.Mensagem) error {
	msgPassagens := Mensagem{}
	err := json.Unmarshal([]byte(mensagem.Dados), &msgPassagens)

	if err != nil {
		return err
	}

	return gormPostgres.Transaction(func(tx *gorm.DB) error {
		if mensagem.Tipo == sagas.EXECUTE {
			switch msgPassagens.Acao {
			case CONFIRMACAO:
				return passagens.Reservar(tx, msgPassagens.IdAssento)
			case TIMEOUT:
				return passagens.Liberar(tx, msgPassagens.IdAssento)
			default:
				return errors.New("ação não definida")
			}
		} else {
			// Não está previsto reverter timeout
			return passagens.ReverterReserva(tx, msgPassagens.IdAssento)
		}
	})
}
