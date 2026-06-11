package puertos

import "context"

type ResumenUsuario struct {
	Identificador string
	NombreCorto   string
	Nombre        string
	Correo        string
	Estado        string
}

type ResumenRol struct {
	Identificador string
	Codigo        string
	Descripcion   string
	EsDeSistema   bool
	Permisos      []string
}

type ResumenPermiso struct {
	Identificador string
	Codigo        string
	Descripcion   string
	Modulo        string
}

type ResumenAsignacion struct {
	Identificador string
	Rol           string
	CodigoRol     string
	AlcanceMina   string
	Vigente       bool
}

type LectorDeGobierno interface {
	ListarUsuarios(ctx context.Context) ([]ResumenUsuario, error)
	ListarRoles(ctx context.Context) ([]ResumenRol, error)
	CatalogoDePermisos(ctx context.Context) ([]ResumenPermiso, error)
	AsignacionesDe(ctx context.Context, identificadorUsuario string) ([]ResumenAsignacion, error)
}
