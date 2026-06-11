package dominio

import "errors"

var (
	ErrCodigoObligatorio     = errors.New("el codigo es obligatorio")
	ErrNombreObligatorio     = errors.New("el nombre es obligatorio")
	ErrCorreoInvalido        = errors.New("el correo no es valido")
	ErrColorDeMarcaInvalido  = errors.New("el color de marca debe ser hexadecimal #RRGGBB")
	ErrMonedaNoSoportada     = errors.New("la moneda no esta soportada")
	ErrRolDeSistemaProtegido = errors.New("un rol de sistema no se puede modificar ni eliminar")
	ErrPermisoYaConcedido    = errors.New("el permiso ya esta concedido al rol")
	ErrPermisoNoConcedido    = errors.New("el permiso no estaba concedido al rol")
	ErrUsuarioYaInactivo     = errors.New("el usuario ya esta inactivo")
	ErrAsignacionYaRevocada  = errors.New("la asignacion de rol ya fue revocada")

	ErrCodigoDeEmpresaObligatorio = errors.New("el codigo de la empresa es obligatorio")
	ErrRazonSocialObligatoria     = errors.New("la razon social es obligatoria")
	ErrUsuarioYaActivo            = errors.New("el usuario ya esta activo")
	ErrEstadoNoReconocido         = errors.New("el estado solicitado no es valido")
)
