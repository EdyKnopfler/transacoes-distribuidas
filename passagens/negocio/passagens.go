package hotel

import (
	"fmt"

	"gorm.io/gorm"
)

const (
	LIVRE       = 1
	PRE_RESERVA = 2
	RESERVADO   = 3
)

type Assento struct {
	ID     string `gorm:"primaryKey;size:36"`
	Estado byte   `gorm:"default:1"`
}

var transicoesValidas = map[byte]map[byte]bool{
	LIVRE: {
		PRE_RESERVA: true,
	},
	PRE_RESERVA: {
		LIVRE:     true,
		RESERVADO: true,
	},
	RESERVADO: {
		PRE_RESERVA: true,
	},
}

func Reservar(db *gorm.DB, idAssento string) error {
	return transicionarEstado(db, idAssento, RESERVADO)
}

func Liberar(db *gorm.DB, idAssento string) error {
	return transicionarEstado(db, idAssento, LIVRE)
}

func ReverterReserva(db *gorm.DB, idAssento string) error {
	return transicionarEstado(db, idAssento, PRE_RESERVA)
}

func transicionarEstado(db *gorm.DB, idAssento string, novoEstado byte) error {
	var assento Assento

	if err := db.First(&assento, idAssento).Error; err != nil {
		return err
	}

	if transicoesValidas[assento.Estado][novoEstado] {
		if err := db.Save(&assento).Error; err != nil {
			return err
		}

		return nil
	} else {
		return fmt.Errorf("transição de assento de voo inválida: %d => %d", assento.Estado, novoEstado)
	}
}
