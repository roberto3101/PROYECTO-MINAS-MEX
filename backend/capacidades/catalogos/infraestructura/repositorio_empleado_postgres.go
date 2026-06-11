package infraestructura

import (
	"context"

	"minas/capacidades/catalogos/dominio"
	"minas/plataforma/persistencia"
)

type RepositorioEmpleadoPostgres struct{}

func NuevoRepositorioEmpleado() RepositorioEmpleadoPostgres {
	return RepositorioEmpleadoPostgres{}
}

func (RepositorioEmpleadoPostgres) Guardar(ctx context.Context, empleado dominio.Empleado) error {
	consultas := persistencia.ConsultasDe(ctx)
	_, err := consultas.Exec(ctx,
		`INSERT INTO catalogos.empleado (id, id_empresa, id_mina, numero_nomina, nombre_completo)
		 VALUES ($1, $2, $3, $4, $5)`,
		empleado.Identificador().Texto(), empleado.Empresa().Texto(), empleado.Mina().Texto(),
		empleado.NumeroNomina(), empleado.NombreCompleto())
	return traducirErrorDeEscritura(err)
}
