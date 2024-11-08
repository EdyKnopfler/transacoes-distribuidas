package main

import (
	"encoding/json"
	"fmt"
)

type Objeto struct {
	Valor     int     `json:"valor"`
	Descricao *string `json:"descricao"`
}

func main() {
	obj := Objeto{}

	err := json.Unmarshal([]byte("{ \"valor\": 1 }"), &obj)
	fmt.Println(err, obj)

	err = json.Unmarshal([]byte("{ \"valor\": 1, \"descricao\": \"Gatim Fofim\" }"), &obj)
	fmt.Println(err, obj, *obj.Descricao)
}
