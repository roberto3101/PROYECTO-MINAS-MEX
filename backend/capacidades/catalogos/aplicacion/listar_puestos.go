package aplicacion

import (
	"context"

	"minas/capacidades/catalogos/puertos"
)

type ListarPuestos struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeCatalogos
}

func NuevoListarPuestos(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeCatalogos) *ListarPuestos {
	return &ListarPuestos{unidad: unidad, lector: lector}
}

func (caso *ListarPuestos) Ejecutar(ctx context.Context) ([]puertos.OpcionDeCatalogo, error) {
	var puestos []puertos.OpcionDeCatalogo
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listados, err := caso.lector.ListarPuestos(ctx)
		puestos = listados
		return err
	})
	return puestos, err
}
