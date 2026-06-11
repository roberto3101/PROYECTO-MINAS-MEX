package aplicacion

import (
	"context"

	"minas/capacidades/catalogos/puertos"
)

type ListarDepartamentos struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeCatalogos
}

func NuevoListarDepartamentos(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeCatalogos) *ListarDepartamentos {
	return &ListarDepartamentos{unidad: unidad, lector: lector}
}

func (caso *ListarDepartamentos) Ejecutar(ctx context.Context) ([]puertos.OpcionDeCatalogo, error) {
	var departamentos []puertos.OpcionDeCatalogo
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listados, err := caso.lector.ListarDepartamentos(ctx)
		departamentos = listados
		return err
	})
	return departamentos, err
}
