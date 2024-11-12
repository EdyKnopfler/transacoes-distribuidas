package main

import (
	"fmt"
	"time"

	"com.derso.aprendendo/conexoes"
	"com.derso.aprendendo/sagas"
	"com.derso.aprendendo/sessoes"
)

const (
	INTERVAL       = 5 * time.Minute // TODO intervalo curto para teste
	RETRY_INTERVAL = 5 * time.Minute
	MAX_RETRIES    = 3
)

var (
	redisSessoes *conexoes.RedisConnection
	redisLock    *conexoes.RedisConnection
	rabbitMQ     *conexoes.RabbitMQConnection
)

func dispararTimeouts() (ok bool) {
	tamanhoLote := int64(50)
	ok = true

	redisSessoes.Iterar(
		tamanhoLote,
		func(idSessao string) error {
			return sessoes.ExecutarSobBloqueio(idSessao, redisLock, func() error {
				var sessao sessoes.Sessao
				err := redisSessoes.ObterObjeto(idSessao, &sessao)

				if err != nil {
					return err
				}

				if !sessao.Expirada() {
					return nil
				}

				var mensagem sagas.Mensagem = map[string]any{
					"tipo":      sagas.DESFAÇA,
					"idSessao":  idSessao,
					"idUsuario": sessao.IdUsuario,
					"idVaga":    sessao.IdVaga,
					"idAssento": sessao.IdAssento,
				}

				// TODO De trás para frente. A ordem de execução deverá ser revisada.
				err = sagas.Publicar(rabbitMQ.Channel, "pagamentos", mensagem)

				if err != nil {
					return err
				}

				return redisSessoes.Deletar(idSessao)
			})
		},
		func(idSessao string, err error) {
			fmt.Printf("Erro ao invalidar sessão '%s': %s", idSessao, err)
			ok = false
		},
	)

	return
}

func main() {
	redisSessoes = conexoes.ConectarRedis(conexoes.SESSION_DATABASE)
	defer redisSessoes.Fechar()

	redisLock = conexoes.ConectarRedis(conexoes.LOCK_DATABASE)
	defer redisLock.Fechar()

	broker, err := conexoes.ConectarRabbitMQ()

	if err != nil {
		panic("Não foi possível conectar-se ao RabbitMQ.")
	}

	rabbitMQ = broker
	defer rabbitMQ.Fechar()

	for {
		result := dispararTimeouts()

		if !result {
			for range MAX_RETRIES {
				time.Sleep(RETRY_INTERVAL)

				if dispararTimeouts() {
					break
				}
			}
		}

		time.Sleep(INTERVAL)
	}
}
