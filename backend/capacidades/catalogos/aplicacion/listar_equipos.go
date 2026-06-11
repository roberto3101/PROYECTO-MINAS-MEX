package aplicacion

import (
	"context"

	"minas/capacidades/catalogos/puertos"
)

type ListarEquipos struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeCatalogos
}

func NuevoListarEquipos(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeCatalogos) *ListarEquipos {
	return &ListarEquipos{unidad: unidad, lector: lector}
}

func (caso *ListarEquipos) Ejecutar(ctx context.Context, filtro puertos.FiltroDeCatalogo) ([]puertos.ResumenEquipo, string, error) {
	var equipos []puertos.ResumenEquipo
	var siguiente string
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listados, cursor, err := caso.lector.ListarEquipos(ctx, filtro)
		equipos, siguiente = listados, cursor
		return err
	})
	return equipos, siguiente, err
}
