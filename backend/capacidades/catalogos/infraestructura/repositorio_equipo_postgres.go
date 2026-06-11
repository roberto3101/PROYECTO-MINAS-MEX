package infraestructura

import (
	"context"

	"minas/capacidades/catalogos/dominio"
	"minas/compartido/identificador"
	"minas/plataforma/persistencia"
)

type RepositorioEquipoPostgres struct{}

func NuevoRepositorioEquipo() RepositorioEquipoPostgres {
	return RepositorioEquipoPostgres{}
}

func (RepositorioEquipoPostgres) Guardar(ctx context.Context, equipo dominio.Equipo) error {
	consultas := persistencia.ConsultasDe(ctx)
	_, err := consultas.Exec(ctx,
		`INSERT INTO catalogos.equipo (id, id_empresa, id_mina, id_tipo_equipo, id_modulo_trabajo, codigo,
		                               descripcion, modelo_perforadora, capacidad_longitud, fabricante, modelo,
		                               numero_serie, anio_fabricacion, fecha_ingreso_mina, estado)
		 VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), NULLIF($8, ''), $9, NULLIF($10, ''), NULLIF($11, ''),
		         NULLIF($12, ''), $13, $14, $15)`,
		equipo.Identificador().Texto(), equipo.Empresa().Texto(), equipo.Mina().Texto(),
		equipo.TipoEquipo().Texto(), equipo.ModuloTrabajo().Texto(), equipo.Codigo(),
		equipo.Descripcion(), equipo.ModeloPerforadora(), equipo.CapacidadLongitud(),
		equipo.Fabricante(), equipo.Modelo(), equipo.NumeroSerie(),
		equipo.AnioFabricacion(), equipo.FechaIngresoMina(), equipo.Estado())
	return traducirErrorDeEscritura(err)
}

func (RepositorioEquipoPostgres) CambiarEstado(ctx context.Context, id identificador.Identificador, estado string) error {
	consultas := persistencia.ConsultasDe(ctx)
	_, err := consultas.Exec(ctx,
		`UPDATE catalogos.equipo SET estado = $2, actualizado_en = now() WHERE id = $1`,
		id.Texto(), estado)
	return err
}
