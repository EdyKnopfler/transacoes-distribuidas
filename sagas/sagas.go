package sagas

import (
	"bytes"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const errorsExchange = "errors_exchange"
const sagasExchange = "sagas"

const EXECUTE byte = 1
const DESFAÇA byte = 2

type Mensagem struct {
	Tipo  byte
	Dados []byte
}

func Inicializar(ch *amqp.Channel) error {
	if err := declararExchange(ch, errorsExchange); err != nil {
		return err
	}

	if err := declararFila(ch, "errors", errorsExchange); err != nil {
		return err
	}

	if err := declararFila(ch, "errors", errorsExchange); err != nil {
		return err
	}

	if err := configurarQos(ch); err != nil {
		return err
	}

	return nil
}

func ConfigurarServicoSagas(
	ch *amqp.Channel,
	filaEsteServico string,
	filaServicoAnterior *string,
	filaProximoServico *string,
) error {
	if err := declararExchange(ch, sagasExchange); err != nil {
		return err
	}

	if filaServicoAnterior != nil {
		if err := declararFila(ch, *filaServicoAnterior, sagasExchange); err != nil {
			return err
		}
	}

	if err := declararFila(ch, filaEsteServico, sagasExchange); err != nil {
		return err
	}

	if filaProximoServico != nil {
		if err := declararFila(ch, *filaProximoServico, sagasExchange); err != nil {
			return err
		}
	}

	return nil
}

func IniciarConsumo(
	ch *amqp.Channel,
	nomeConsumidor string,
	funcaoTratamento func(Mensagem) error,
	filaEsteServico string,
	filaServicoAnterior *string,
	filaProximoServico *string,
) error {
	entregas, err := ch.Consume(
		filaEsteServico,
		nomeConsumidor,
		false, // auto ack
		false, // exclusive
		false, // "no local" não suportado pelo RabbitMQ
		false, // no wait
		nil,
	)

	if err != nil {
		fmt.Printf("Falha ao iniciar consumo na fila '%s' no RabbitMQ: %s\n", filaEsteServico, err)
		return err
	}

	var mensagem Mensagem

	for entrega := range entregas {
		buffer := bytes.NewBuffer(entrega.Body)
		decoder := json.NewDecoder(buffer)
		err := decoder.Decode(&mensagem)

		if err != nil {
			fmt.Printf("Falha ao ler mensagem '%s' em JSON: %s\n", buffer.String(), err)
		} else {
			err = funcaoTratamento(mensagem)
		}

		if err != nil {
			// multiple, requeue
			if err = entrega.Nack(false, false); err != nil {
				fmt.Printf("Falha ao realizar Not Ack na fila '%s' no RabbitMQ: %s\n", filaEsteServico, err)
			}

			if filaServicoAnterior != nil {
				mensagem.Tipo = DESFAÇA
				Publicar(ch, *filaServicoAnterior, mensagem)
			}
		} else {
			// multiple
			if err = entrega.Ack(false); err != nil {
				fmt.Printf("Falha ao realizar Ack na fila '%s' no RabbitMQ: %s\n", filaEsteServico, err)
			}

			var fila *string
			if mensagem.Tipo == EXECUTE {
				fila = filaProximoServico
			} else {
				fila = filaServicoAnterior
			}

			if fila != nil {
				Publicar(ch, *fila, mensagem)
			}
		}
	}

	return nil // Só por burocracia
}

func Publicar(ch *amqp.Channel, fila string, mensagem Mensagem) error {
	// Escolhido formato JSON devido à troca contínua de inúmeras mensagens pequenas
	// https://rsheremeta.medium.com/benchmarking-gob-vs-json-xml-yaml-48b090b097e8W
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	err := encoder.Encode(mensagem)

	if err != nil {
		fmt.Printf(
			"Falha ao transformar mensagem tipo %d: '%s' em JSON: %s\n",
			mensagem.Tipo,
			mensagem.Dados,
			err,
		)
		return err
	}

	err = ch.Publish(
		sagasExchange, // exchange
		fila,          // routing key
		true,          // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        buffer.Bytes(),
		},
	)

	if err != nil {
		fmt.Printf(
			"Falha ao publicar mensagem tipo %d: '%s' no RabbitMQ: %s\n",
			mensagem.Tipo,
			mensagem.Dados,
			err,
		)
		return err
	}

	return nil
}

func declararExchange(ch *amqp.Channel, nome string) error {
	err := ch.ExchangeDeclare(
		nome,
		"direct", // type
		true,     // durable
		false,    // autodelete
		false,    // internal
		false,    // no wait
		nil,
	)

	if err != nil {
		fmt.Printf("Falha ao declarar exchange '%s' no RabbitMQ: %s\n", nome, err)
		return err
	}

	return nil
}

func declararFila(ch *amqp.Channel, nomeFila, exchange string) error {
	_, err := ch.QueueDeclare(
		nomeFila,
		true,  // durable
		false, // autodelete
		false, // exclusive
		false, // no wait
		amqp.Table{
			"x-dead-letter-exchange":    "errors",
			"x-dead-letter-routing-key": "errors",
		},
	)

	if err != nil {
		fmt.Printf("Falha ao declarar fila '%s' no RabbitMQ: %s\n", nomeFila, err)
		return err
	}

	err = ch.QueueBind(
		nomeFila,
		nomeFila, // routing key
		exchange,
		false, // no wait
		nil,
	)

	if err != nil {
		fmt.Printf("Falha ao realizar binding da fila '%s' com o exchange no RabbitMQ: %s\n", nomeFila, err)
		return err
	}

	return nil
}

func configurarQos(ch *amqp.Channel) error {
	err := ch.Qos(
		0,     // prefetch count
		1,     // prefetch size
		false, // global
	)

	if err != nil {
		fmt.Printf("Falha ao configurar QoS no RabbitMQ: %s\n", err)
		return err
	}

	return nil
}
