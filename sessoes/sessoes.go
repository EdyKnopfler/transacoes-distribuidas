package sessoes

import (
	"fmt"
	"time"

	"com.derso.aprendendo/conexoes"
	"github.com/google/uuid"
)

const (
	ATIVA      = 1
	FINALIZADA = 2
	CANCELADA  = 3
	EXPIRA_EM  = 2 * time.Minute // TODO Valor baixo para testes
)

type Sessao struct {
	IdUsuario string `redis:"idUsuario" json:"idUsuario"`
	IdVaga    string `redis:"idVaga" json:"idVaga"`
	IdAssento string `redis:"idAssento" json:"idAssento"`
	Criacao   int64  `redis:"criacao" json:"criacao"`
	Estado    byte   `redis:"estado" json:"estado"`
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

func CriarSessao(redis *conexoes.RedisConnection, sessao Sessao) (string, error) {
	idSessao := uuid.New().String()
	criacao := time.Now().Unix()

	err := redis.SetarObjeto(
		idSessao,
		"idUsuario", sessao.IdUsuario,
		"idVaga", sessao.IdVaga,
		"idAssento", sessao.IdAssento,
		"criacao", criacao,
		"estado", ATIVA,
	)

	if err != nil {
		return "", err
	}

	return idSessao, nil
}

func MudarEstado(idSessao string, redis *conexoes.RedisConnection, novoEstado byte) error {
	var sessao Sessao
	err := redis.ObterObjeto(idSessao, &sessao)

	if err != nil {
		return err
	}

	estadoAnterior := sessao.Estado

	if transicoesValidas[estadoAnterior][novoEstado] {
		return redis.SetarObjeto(
			idSessao,
			"idUsuario", sessao.IdUsuario,
			"idVaga", sessao.IdVaga,
			"idAssento", sessao.IdAssento,
			"estado", novoEstado,
		)
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

func (sessao *Sessao) Expirada() bool {
	return time.Now().After(time.Unix(sessao.Criacao, 0).Add(EXPIRA_EM))
}
