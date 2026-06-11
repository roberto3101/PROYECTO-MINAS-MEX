package puertos

import "context"

type FiltroDeCatalogo struct {
	Busqueda string
	Estado   string
	Cursor   string
	Limite   int
}

type ResumenMina struct {
	Identificador string
	Nombre        string
	Area          string
	Niveles       string
	Estado        string
}

type DetalleMina struct {
	ResumenMina
	DensidadPromedio string
	SeccionObraLargo string
	SeccionObraAncho string
	SeccionVetaLargo string
	SeccionVetaAncho string
	CreadoEn         string
}

type ResumenEmpleado struct {
	Identificador  string
	NumeroNomina   string
	NombreCompleto string
	Mina           string
	Departamento   string
	Puesto         string
	Grupo          string
	Estado         string
}

type DetalleEmpleado struct {
	ResumenEmpleado
	Actividad     string
	CentroCosto   string
	GerenteACargo string
	CreadoEn      string
}

type ResumenEquipo struct {
	Identificador string
	Codigo        string
	Tipo          string
	Modulo        string
	Mina          string
	Fabricante    string
	Modelo        string
	Estado        string
}

type DetalleEquipo struct {
	ResumenEquipo
	Descripcion       string
	NumeroSerie       string
	FechaIngresoMina  string
	ModeloPerforadora string
	CapacidadLongitud string
	CreadoEn          string
	AnioFabricacion   int
}

type OpcionDeCatalogo struct {
	Identificador string
	Codigo        string
	Descripcion   string
}

type LectorDeCatalogos interface {
	ListarMinas(ctx context.Context, filtro FiltroDeCatalogo) ([]ResumenMina, string, error)
	DetalleDeMina(ctx context.Context, identificadorMina string) (DetalleMina, bool, error)
	ListarEmpleados(ctx context.Context, filtro FiltroDeCatalogo) ([]ResumenEmpleado, string, error)
	DetalleDeEmpleado(ctx context.Context, identificadorEmpleado string) (DetalleEmpleado, bool, error)
	ListarEquipos(ctx context.Context, filtro FiltroDeCatalogo) ([]ResumenEquipo, string, error)
	DetalleDeEquipo(ctx context.Context, identificadorEquipo string) (DetalleEquipo, bool, error)
	ListarTiposDeEquipo(ctx context.Context) ([]OpcionDeCatalogo, error)
	ListarModulosDeTrabajo(ctx context.Context) ([]OpcionDeCatalogo, error)
	ListarDepartamentos(ctx context.Context) ([]OpcionDeCatalogo, error)
	ListarPuestos(ctx context.Context) ([]OpcionDeCatalogo, error)
	ListarActividades(ctx context.Context) ([]OpcionDeCatalogo, error)
}
