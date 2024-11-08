package sessoes

import (
	"context"
	"encoding/json"

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
	Estado    int
}

func CriarSessao(ctx context.Context, redis conexoes.RedisConnection, IdUsuario string) (string, error) {
	id := uuid.New().String()

	sessao := Sessao{
		IdUsuario: IdUsuario,
		Estado:    ATIVA,
	}

	sessaoJson, _ := json.Marshal(sessao)

	err := redis.Setar(ctx, id, string(sessaoJson))

	if err != nil {
		return id, err
	}

	return id, nil
}
