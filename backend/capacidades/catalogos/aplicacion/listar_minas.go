package aplicacion

import (
	"context"

	"minas/capacidades/catalogos/puertos"
)

type ListarMinas struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeCatalogos
}

func NuevoListarMinas(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeCatalogos) *ListarMinas {
	return &ListarMinas{unidad: unidad, lector: lector}
}

func (caso *ListarMinas) Ejecutar(ctx context.Context) ([]puertos.ResumenMina, error) {
	var minas []puertos.ResumenMina
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listadas, err := caso.lector.ListarMinas(ctx)
		minas = listadas
		return err
	})
	return minas, err
}
