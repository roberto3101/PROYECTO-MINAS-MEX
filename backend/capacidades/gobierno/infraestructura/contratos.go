package infraestructura

import (
	"minas/capacidades/gobierno/contrato"
	"minas/capacidades/gobierno/puertos"
)

var (
	_ puertos.RepositorioUsuario       = RepositorioUsuarioPostgres{}
	_ puertos.RepositorioRol           = RepositorioRolPostgres{}
	_ puertos.RepositorioAsignacionRol = RepositorioAsignacionPostgres{}
	_ puertos.RepositorioEmpresa       = RepositorioEmpresaPostgres{}
	_ puertos.RepositorioPermiso       = RepositorioPermisoPostgres{}
	_ puertos.LectorDeAcceso           = LectorDeAccesoPostgres{}
	_ contrato.Gobierno                = ServicioGobiernoPostgres{}
)
