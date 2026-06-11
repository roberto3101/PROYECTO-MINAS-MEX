package aplicacion

import (
	"context"

	"minas/capacidades/catalogos/puertos"
)

type ListarActividades struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeCatalogos
}

func NuevoListarActividades(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeCatalogos) *ListarActividades {
	return &ListarActividades{unidad: unidad, lector: lector}
}

func (caso *ListarActividades) Ejecutar(ctx context.Context) ([]puertos.OpcionDeCatalogo, error) {
	var actividades []puertos.OpcionDeCatalogo
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listadas, err := caso.lector.ListarActividades(ctx)
		actividades = listadas
		return err
	})
	return actividades, err
}
