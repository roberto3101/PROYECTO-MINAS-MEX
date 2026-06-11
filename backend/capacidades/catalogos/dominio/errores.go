package dominio

import "errors"

var (
	ErrNombreObligatorio        = errors.New("el nombre es obligatorio")
	ErrCodigoObligatorio        = errors.New("el codigo es obligatorio")
	ErrNumeroNominaObligatorio  = errors.New("el numero de nomina es obligatorio")
	ErrMinaObligatoria          = errors.New("la mina es obligatoria")
	ErrTipoDeEquipoObligatorio  = errors.New("el tipo de equipo es obligatorio")
	ErrModuloTrabajoObligatorio = errors.New("el modulo de trabajo es obligatorio")
	ErrYaExiste                 = errors.New("ya existe un registro activo con esa clave")
)
