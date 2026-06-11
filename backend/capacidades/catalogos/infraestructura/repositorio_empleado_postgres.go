package infraestructura

import (
	"context"

	"minas/capacidades/catalogos/dominio"
	"minas/compartido/identificador"
	"minas/plataforma/persistencia"
)

type RepositorioEmpleadoPostgres struct{}

func NuevoRepositorioEmpleado() RepositorioEmpleadoPostgres {
	return RepositorioEmpleadoPostgres{}
}

func (RepositorioEmpleadoPostgres) Guardar(ctx context.Context, empleado dominio.Empleado) error {
	consultas := persistencia.ConsultasDe(ctx)
	_, err := consultas.Exec(ctx,
		`INSERT INTO catalogos.empleado (id, id_empresa, id_mina, numero_nomina, nombre_completo,
		                                 id_departamento, id_puesto, id_actividad, centro_costo, gerente_a_cargo, grupo)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9, ''), NULLIF($10, ''), NULLIF($11, ''))`,
		empleado.Identificador().Texto(), empleado.Empresa().Texto(), empleado.Mina().Texto(),
		empleado.NumeroNomina(), empleado.NombreCompleto(),
		textoOpcional(empleado.Departamento()), textoOpcional(empleado.Puesto()), textoOpcional(empleado.Actividad()),
		empleado.CentroCosto(), empleado.GerenteACargo(), empleado.Grupo())
	return traducirErrorDeEscritura(err)
}

func (RepositorioEmpleadoPostgres) CambiarEstado(ctx context.Context, id identificador.Identificador, estado string) error {
	consultas := persistencia.ConsultasDe(ctx)
	_, err := consultas.Exec(ctx,
		`UPDATE catalogos.empleado SET estado = $2, actualizado_en = now() WHERE id = $1`,
		id.Texto(), estado)
	return err
}
