package puertos

import "context"

type FiltroDeUsuarios struct {
	Busqueda string
	Estado   string
	Cursor   string
	Limite   int
}

type ResumenUsuario struct {
	Identificador  string
	NombreCorto    string
	Nombre         string
	Correo         string
	Estado         string
	UltimoAccesoEn string
}

type DetalleUsuario struct {
	ResumenUsuario
	CreadoEn              string
	IdentificadorEmpleado string
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

type MinaConAcceso struct {
	Identificador string
	Nombre        string
}

type AccesoAMinas struct {
	EsGlobal bool
	Minas    []MinaConAcceso
}

type LectorDeGobierno interface {
	ListarUsuarios(ctx context.Context, filtro FiltroDeUsuarios) ([]ResumenUsuario, string, error)
	DetalleDeUsuario(ctx context.Context, identificadorUsuario string) (DetalleUsuario, bool, error)
	ListarRoles(ctx context.Context) ([]ResumenRol, error)
	CatalogoDePermisos(ctx context.Context) ([]ResumenPermiso, error)
	AsignacionesDe(ctx context.Context, identificadorUsuario string) ([]ResumenAsignacion, error)
	MinasAccesibles(ctx context.Context, identificadorUsuario string) (AccesoAMinas, error)
}
