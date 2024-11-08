package sessoes

import (
	"encoding/json"
	"fmt"

	"com.derso.aprendendo/conexoes"
	"github.com/google/uuid"
)

const (
	ATIVA      = 1
	FINALIZADA = 2
	CANCELADA  = 3
)

type Sessao struct {
	IdUsuario string
	Estado    byte
}

var transicoesValidas = map[byte]map[byte]bool{
	ATIVA: {
		FINALIZADA: true,
		CANCELADA:  true,
	},
	FINALIZADA: {
		ATIVA: true, // Em caso de reversão do processo de finalização
	},
	CANCELADA: {},
}

func CriarSessao(redis conexoes.RedisConnection, IdUsuario string) (string, error) {
	id := uuid.New().String()

	sessao := Sessao{
		IdUsuario: IdUsuario,
		Estado:    ATIVA,
	}

	sessaoJson, err := json.Marshal(sessao)

	if err != nil {
		return id, err
	}

	err = redis.Setar(id, string(sessaoJson))

	if err != nil {
		return id, err
	}

	return id, nil
}

func MudarEstado(idSessao string, redis *conexoes.RedisConnection, novoEstado byte) error {
	raw, err := redis.Obter(idSessao)

	if err != nil {
		return err
	}

	var sessao Sessao
	err = json.Unmarshal([]byte(raw), &sessao)

	if err != nil {
		return err
	}

	estadoAnterior := sessao.Estado

	if transicoesValidas[estadoAnterior][novoEstado] {
		sessao.Estado = novoEstado
		sessaoJson, err := json.Marshal(sessao)

		if err != nil {
			return err
		}

		err = redis.Setar(idSessao, string(sessaoJson))
		return err
	} else {
		return fmt.Errorf("transição de sessão inválida: %d => %d", estadoAnterior, novoEstado)
	}
}

func ExecutarSobBloqueio(idSessao string, redis *conexoes.RedisConnection, fn func() error) error {
	mutex, err := redis.Bloquear(idSessao)

	if err != nil {
		return err
	}

	defer redis.Desbloquear(mutex)

	return fn()
}
