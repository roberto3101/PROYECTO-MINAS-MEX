package infraestructura

import (
	"context"

	"minas/capacidades/catalogos/dominio"
	"minas/plataforma/persistencia"
)

type RepositorioEquipoPostgres struct{}

func NuevoRepositorioEquipo() RepositorioEquipoPostgres {
	return RepositorioEquipoPostgres{}
}

func (RepositorioEquipoPostgres) Guardar(ctx context.Context, equipo dominio.Equipo) error {
	consultas := persistencia.ConsultasDe(ctx)
	_, err := consultas.Exec(ctx,
		`INSERT INTO catalogos.equipo (id, id_empresa, id_mina, id_tipo_equipo, id_modulo_trabajo, codigo, fabricante)
		 VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''))`,
		equipo.Identificador().Texto(), equipo.Empresa().Texto(), equipo.Mina().Texto(),
		equipo.TipoEquipo().Texto(), equipo.ModuloTrabajo().Texto(), equipo.Codigo(), equipo.Fabricante())
	return traducirErrorDeEscritura(err)
}
