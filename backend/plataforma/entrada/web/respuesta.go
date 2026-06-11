package web

import (
	"encoding/json"
	"errors"
	"net/http"
)

const tamanoMaximoDeCuerpo = 64 << 10

func ResponderJson(escritor http.ResponseWriter, codigo int, cuerpo any) {
	escritor.Header().Set("Content-Type", "application/json; charset=utf-8")
	escritor.WriteHeader(codigo)
	_ = json.NewEncoder(escritor).Encode(cuerpo)
}

func ResponderError(escritor http.ResponseWriter, codigo int, mensaje string) {
	ResponderJson(escritor, codigo, map[string]string{"error": mensaje})
}

func DecodificarCuerpo(escritor http.ResponseWriter, peticion *http.Request, destino any) bool {
	peticion.Body = http.MaxBytesReader(escritor, peticion.Body, tamanoMaximoDeCuerpo)
	decodificador := json.NewDecoder(peticion.Body)
	decodificador.DisallowUnknownFields()
	if err := decodificador.Decode(destino); err != nil {
		var demasiadoGrande *http.MaxBytesError
		if errors.As(err, &demasiadoGrande) {
			ResponderError(escritor, http.StatusRequestEntityTooLarge, "el cuerpo de la peticion es demasiado grande")
			return false
		}
		ResponderError(escritor, http.StatusBadRequest, "el cuerpo de la peticion no es valido o trae campos desconocidos")
		return false
	}
	return true
}

func ConCabecerasDeSeguridad(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(escritor http.ResponseWriter, peticion *http.Request) {
		cabeceras := escritor.Header()
		cabeceras.Set("X-Content-Type-Options", "nosniff")
		cabeceras.Set("X-Frame-Options", "DENY")
		cabeceras.Set("Referrer-Policy", "no-referrer")
		cabeceras.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		cabeceras.Set("Content-Security-Policy",
			"default-src 'self'; img-src 'self' data: https:; style-src 'self' 'unsafe-inline'; script-src 'self'; object-src 'none'; frame-ancestors 'none'")
		siguiente.ServeHTTP(escritor, peticion)
	})
}
