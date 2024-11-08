package hotel

import "gorm.io/gorm"

const (
	LIVRE       = 1
	PRE_RESERVA = 2
	RESERVADA   = 3
)

type Vaga struct {
	ID     string `gorm:"primaryKey;size:36"`
	Estado int
}

func Confirmar(db *gorm.DB, idVaga string) error {
	// TODO Lógica com o gORM: verificar o estado atual da vaga
	return nil
}

func Cancelar(db *gorm.DB, idVaga string) error {
	return nil
}
