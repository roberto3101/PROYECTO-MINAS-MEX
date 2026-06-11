package puertos

import "context"

type ResumenMina struct {
	Identificador string
	Nombre        string
	Area          string
}

type ResumenEmpleado struct {
	Identificador  string
	NumeroNomina   string
	NombreCompleto string
	Mina           string
}

type ResumenEquipo struct {
	Identificador string
	Codigo        string
	Mina          string
	Tipo          string
	Modulo        string
	Fabricante    string
}

type OpcionDeCatalogo struct {
	Identificador string
	Codigo        string
	Descripcion   string
}

type LectorDeCatalogos interface {
	ListarMinas(ctx context.Context) ([]ResumenMina, error)
	ListarEmpleados(ctx context.Context) ([]ResumenEmpleado, error)
	ListarEquipos(ctx context.Context) ([]ResumenEquipo, error)
	ListarTiposDeEquipo(ctx context.Context) ([]OpcionDeCatalogo, error)
	ListarModulosDeTrabajo(ctx context.Context) ([]OpcionDeCatalogo, error)
}
