package sagas

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const errorsExchange = "errors_exchange"
const errorsRoutingKey = "errors"
const sagasExchange = "sagas"

const EXECUTE float64 = 1
const DESFAÇA float64 = 2

type Mensagem = map[string]any

func Inicializar(ch *amqp.Channel) error {
	if err := declararExchange(ch, errorsExchange); err != nil {
		return err
	}

	if err := declararFila(ch, errorsRoutingKey, errorsExchange); err != nil {
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

	for entrega := range entregas {
		mensagem, err := decodeMsg(entrega.Body)

		if err != nil {
			fmt.Printf("Falha ao ler mensagem '%s' em JSON: %s\n", string(entrega.Body), err)
		} else {
			_, presente := mensagem["tipo"]

			if !presente {
				mensagem["tipo"] = EXECUTE
			}

			err = tratarErro(funcaoTratamento, mensagem)
		}

		if err != nil {
			fmt.Printf("Erro ao processar mensagem: %s\n", err)

			// multiple, requeue
			if err = entrega.Nack(false, false); err != nil {
				fmt.Printf("Falha ao realizar Not Ack na fila '%s' no RabbitMQ: %s\n", filaEsteServico, err)
			}

			if filaServicoAnterior != nil {
				mensagem["tipo"] = DESFAÇA
				Publicar(ch, *filaServicoAnterior, mensagem)
			}
		} else {
			// multiple
			if err = entrega.Ack(false); err != nil {
				fmt.Printf("Falha ao realizar Ack na fila '%s' no RabbitMQ: %s\n", filaEsteServico, err)
			}

			tipo, presente := mensagem["tipo"]

			if !presente {
				tipo = EXECUTE
			}

			var fila *string

			if tipo == EXECUTE {
				fila = filaProximoServico
			} else {
				fila = filaServicoAnterior
			}

			if fila != nil {
				Publicar(ch, *fila, mensagem)
			}
		}
	}

	fmt.Println("Finalizando consumidor " + nomeConsumidor)
	return nil
}

func Publicar(ch *amqp.Channel, fila string, mensagem interface{}) error {
	// Escolhido formato JSON devido à troca contínua de inúmeras mensagens pequenas
	// https://rsheremeta.medium.com/benchmarking-gob-vs-json-xml-yaml-48b090b097e8W
	msgBytes, err := json.Marshal(mensagem)

	if err != nil {
		fmt.Printf("Falha ao transformar mensagem '%s' em JSON:\n%s\n\n", mensagem, err)
		return err
	}

	err = ch.Publish(
		sagasExchange, // exchange
		fila,          // routing key
		true,          // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msgBytes,
		},
	)

	if err != nil {
		fmt.Printf("Falha ao publicar mensagem '%s':\n%s\n\n", mensagem, err)
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
			"x-dead-letter-exchange":    errorsExchange,
			"x-dead-letter-routing-key": errorsRoutingKey,
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
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)

	if err != nil {
		fmt.Printf("Falha ao configurar QoS no RabbitMQ: %s\n", err)
		return err
	}

	return nil
}

func decodeMsg(corpo []byte) (mensagem Mensagem, err error) {
	var msgRaw interface{}

	if err = json.Unmarshal(corpo, &msgRaw); err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("estrutura do JSON inválida: %s", r)
		}
	}()

	mensagem = msgRaw.(Mensagem)
	return
}

func tratarErro(funcaoTratamento func(Mensagem) error, mensagem Mensagem) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("erro ao tratar mensagem: %s", r)
		}
	}()

	err = funcaoTratamento(mensagem)
	return
}
