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

func (caso *ListarMinas) Ejecutar(ctx context.Context, filtro puertos.FiltroDeCatalogo) ([]puertos.ResumenMina, string, error) {
	var minas []puertos.ResumenMina
	var siguiente string
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listadas, cursor, err := caso.lector.ListarMinas(ctx, filtro)
		minas, siguiente = listadas, cursor
		return err
	})
	return minas, siguiente, err
}
