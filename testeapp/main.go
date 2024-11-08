package main

import (
	"encoding/json"
	"fmt"
)

type Objeto struct {
	Valor     int     `json:"valor"`
	Descricao *string `json:"descricao"`
}

const (
	LIVRE       = 1
	PRE_RESERVA = 2
	RESERVADA   = 3
)

var transicoesValidas = map[byte]map[byte]bool{
	LIVRE: map[byte]bool{
		PRE_RESERVA: true,
	},
	PRE_RESERVA: map[byte]bool{
		LIVRE:     true,
		RESERVADA: true,
	},
	RESERVADA: map[byte]bool{
		PRE_RESERVA: true,
	},
}

func main() {
	obj := Objeto{}

	err := json.Unmarshal([]byte("{ \"valor\": 1 }"), &obj)
	fmt.Println(err, obj)

	err = json.Unmarshal([]byte("{ \"valor\": 1, \"descricao\": \"Gatim Fofim\" }"), &obj)
	fmt.Println(err, obj, *obj.Descricao)

	fmt.Println(transicoesValidas[LIVRE], transicoesValidas[LIVRE][PRE_RESERVA], transicoesValidas[LIVRE][RESERVADA])

}
