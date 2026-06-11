package aplicacion

import (
	"context"

	"minas/capacidades/catalogos/puertos"
)

type ListarModulosDeTrabajo struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeCatalogos
}

func NuevoListarModulosDeTrabajo(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeCatalogos) *ListarModulosDeTrabajo {
	return &ListarModulosDeTrabajo{unidad: unidad, lector: lector}
}

func (caso *ListarModulosDeTrabajo) Ejecutar(ctx context.Context) ([]puertos.OpcionDeCatalogo, error) {
	var modulos []puertos.OpcionDeCatalogo
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listados, err := caso.lector.ListarModulosDeTrabajo(ctx)
		modulos = listados
		return err
	})
	return modulos, err
}
