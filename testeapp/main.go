package main

import (
	"encoding/json"
	"fmt"

	hotel "com.derso.aprendendo/hotel/negocio"
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

func main() {
	obj := Objeto{}

	err := json.Unmarshal([]byte("{ \"valor\": 1 }"), &obj)
	fmt.Println(err, obj)

	err = json.Unmarshal([]byte("{ \"valor\": 1, \"descricao\": \"Gatim Fofim\" }"), &obj)
	fmt.Println(err, obj, *obj.Descricao)

	fmt.Println(transicoesValidas[LIVRE], transicoesValidas[LIVRE][PRE_RESERVA], transicoesValidas[LIVRE][RESERVADA])

	var arbitrario interface{}

	err = json.Unmarshal([]byte(`
		{ 
			"id": 1, 
			"itens": [
				{ "id": 10, "nome": "Gatim Fofim" }, 
				{ "id": 20, "nome": "Xorrim do Fucim Bunitim" }
			]
		}
	`), &arbitrario)
	fmt.Println(err, arbitrario)

	mapa := arbitrario.(map[string]any)
	fmt.Println("Itens:", mapa["itens"]) // existente
	fmt.Println("Valor:", mapa["valor"]) // inexistente

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Não conseguiu ler:", r)
		}
	}()

	var vaga interface{} = hotel.Vaga{}
	mapa = vaga.(map[string]any) // inválido
	fmt.Println("Conseguiu ler:", mapa)

	fmt.Println("Continuando após recover...")
}
