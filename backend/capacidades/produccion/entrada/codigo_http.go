package entrada

import (
	"errors"
	"net/http"

	"minas/capacidades/produccion/dominio"
)

func codigoHttp(err error) int {
	switch {
	case errors.Is(err, dominio.ErrParteNoEncontrado):
		return http.StatusNotFound
	default:
		return http.StatusBadRequest
	}
}
