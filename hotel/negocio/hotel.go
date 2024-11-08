package hotel

import (
	"fmt"

	"gorm.io/gorm"
)

const (
	LIVRE       = 1
	PRE_RESERVA = 2
	RESERVADA   = 3
)

type Vaga struct {
	ID     string `gorm:"primaryKey;size:36"`
	Estado byte   `gorm:"default:1"`
}

var transicoesValidas = map[byte]map[byte]bool{
	LIVRE: {
		PRE_RESERVA: true,
	},
	PRE_RESERVA: {
		LIVRE:     true,
		RESERVADA: true,
	},
	RESERVADA: {
		PRE_RESERVA: true,
	},
}

func Reservar(db *gorm.DB, idVaga string) error {
	return transicionarEstado(db, idVaga, RESERVADA)
}

func Liberar(db *gorm.DB, idVaga string) error {
	return transicionarEstado(db, idVaga, LIVRE)
}

func ReverterReserva(db *gorm.DB, idVaga string) error {
	return transicionarEstado(db, idVaga, PRE_RESERVA)
}

func transicionarEstado(db *gorm.DB, idVaga string, novoEstado byte) error {
	var vaga Vaga

	if err := db.First(&vaga, idVaga).Error; err != nil {
		return err
	}

	if transicoesValidas[vaga.Estado][novoEstado] {
		if err := db.Save(&vaga).Error; err != nil {
			return err
		}

		return nil
	} else {
		return fmt.Errorf("transição de vaga de hotel inválida: %d => %d", vaga.Estado, novoEstado)
	}
}
