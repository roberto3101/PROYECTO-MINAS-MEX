package aplicacion

import "errors"

var (
	ErrUsuarioDuplicado       = errors.New("ya existe un usuario activo con ese nombre corto")
	ErrUsuarioNoEncontrado    = errors.New("usuario no encontrado")
	ErrRolNoEncontrado        = errors.New("rol no encontrado")
	ErrPermisoNoEncontrado    = errors.New("permiso no encontrado")
	ErrAsignacionNoEncontrada = errors.New("asignacion de rol no encontrada")
	ErrEmpresaNoEncontrada    = errors.New("empresa no encontrada")
)
