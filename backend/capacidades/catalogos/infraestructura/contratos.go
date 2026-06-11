package infraestructura

import (
	"minas/capacidades/catalogos/contrato"
	"minas/capacidades/catalogos/puertos"
)

var (
	_ puertos.RepositorioMina     = RepositorioMinaPostgres{}
	_ puertos.RepositorioEmpleado = RepositorioEmpleadoPostgres{}
	_ puertos.RepositorioEquipo   = RepositorioEquipoPostgres{}
	_ puertos.LectorDeCatalogos   = LectorDeCatalogosPostgres{}
	_ contrato.Catalogos          = ServicioCatalogosPostgres{}
)
