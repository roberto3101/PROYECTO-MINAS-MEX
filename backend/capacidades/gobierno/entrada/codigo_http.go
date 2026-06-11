package entrada

import (
	"errors"
	"net/http"

	"minas/capacidades/gobierno/aplicacion"
	"minas/capacidades/gobierno/dominio"
)

func codigoHttp(err error) int {
	switch {
	case errors.Is(err, aplicacion.ErrUsuarioNoEncontrado),
		errors.Is(err, aplicacion.ErrRolNoEncontrado),
		errors.Is(err, aplicacion.ErrPermisoNoEncontrado),
		errors.Is(err, aplicacion.ErrAsignacionNoEncontrada),
		errors.Is(err, aplicacion.ErrEmpresaNoEncontrada):
		return http.StatusNotFound
	case errors.Is(err, aplicacion.ErrUsuarioDuplicado),
		errors.Is(err, dominio.ErrPermisoYaConcedido),
		errors.Is(err, dominio.ErrPermisoNoConcedido),
		errors.Is(err, dominio.ErrUsuarioYaInactivo),
		errors.Is(err, dominio.ErrAsignacionYaRevocada):
		return http.StatusConflict
	case errors.Is(err, aplicacion.ErrCredencialesInvalidas):
		return http.StatusUnauthorized
	case errors.Is(err, dominio.ErrRolDeSistemaProtegido):
		return http.StatusForbidden
	default:
		return http.StatusBadRequest
	}
}
