package aplicacion

import (
	"context"

	"minas/capacidades/catalogos/puertos"
)

type ListarTiposDeEquipo struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeCatalogos
}

func NuevoListarTiposDeEquipo(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeCatalogos) *ListarTiposDeEquipo {
	return &ListarTiposDeEquipo{unidad: unidad, lector: lector}
}

func (caso *ListarTiposDeEquipo) Ejecutar(ctx context.Context) ([]puertos.OpcionDeCatalogo, error) {
	var tipos []puertos.OpcionDeCatalogo
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listados, err := caso.lector.ListarTiposDeEquipo(ctx)
		tipos = listados
		return err
	})
	return tipos, err
}
