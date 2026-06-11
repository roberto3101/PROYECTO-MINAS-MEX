package entrada

import (
	"errors"
	"net/http"

	"minas/capacidades/catalogos/dominio"
)

func codigoHttp(err error) int {
	if errors.Is(err, dominio.ErrYaExiste) {
		return http.StatusConflict
	}
	return http.StatusBadRequest
}
