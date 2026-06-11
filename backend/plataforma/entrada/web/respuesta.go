package web

import (
	"encoding/json"
	"net/http"
)

func ResponderJson(escritor http.ResponseWriter, codigo int, cuerpo any) {
	escritor.Header().Set("Content-Type", "application/json; charset=utf-8")
	escritor.WriteHeader(codigo)
	_ = json.NewEncoder(escritor).Encode(cuerpo)
}

func ResponderError(escritor http.ResponseWriter, codigo int, mensaje string) {
	ResponderJson(escritor, codigo, map[string]string{"error": mensaje})
}

func DecodificarCuerpo(escritor http.ResponseWriter, peticion *http.Request, destino any) bool {
	if err := json.NewDecoder(peticion.Body).Decode(destino); err != nil {
		ResponderError(escritor, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return false
	}
	return true
}
