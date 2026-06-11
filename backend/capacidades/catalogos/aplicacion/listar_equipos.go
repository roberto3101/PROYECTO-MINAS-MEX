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

func (caso *ListarEquipos) Ejecutar(ctx context.Context) ([]puertos.ResumenEquipo, error) {
	var equipos []puertos.ResumenEquipo
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listados, err := caso.lector.ListarEquipos(ctx)
		equipos = listados
		return err
	})
	return equipos, err
}
