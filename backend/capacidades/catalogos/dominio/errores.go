package dominio

import "errors"

var (
	ErrNombreObligatorio        = errors.New("el nombre es obligatorio")
	ErrAreaObligatoria          = errors.New("el area es obligatoria")
	ErrCodigoObligatorio        = errors.New("el codigo es obligatorio")
	ErrNumeroNominaObligatorio  = errors.New("el numero de nomina es obligatorio")
	ErrMinaObligatoria          = errors.New("la mina es obligatoria")
	ErrTipoDeEquipoObligatorio  = errors.New("el tipo de equipo es obligatorio")
	ErrModuloTrabajoObligatorio = errors.New("el modulo de trabajo es obligatorio")
	ErrValorFueraDeRango        = errors.New("el valor esta fuera del rango permitido")
	ErrEstadoNoReconocido       = errors.New("el estado no es reconocido")
	ErrYaExiste                 = errors.New("ya existe un registro activo con esa clave")
)

func EsEstadoValido(estado string, validos []string) bool {
	for _, valido := range validos {
		if estado == valido {
			return true
		}
	}
	return false
}
