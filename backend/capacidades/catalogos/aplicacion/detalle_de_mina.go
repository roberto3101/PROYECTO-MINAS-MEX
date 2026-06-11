package aplicacion

import (
	"context"

	"minas/capacidades/catalogos/puertos"
)

type DetalleDeMina struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeCatalogos
}

func NuevoDetalleDeMina(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeCatalogos) *DetalleDeMina {
	return &DetalleDeMina{unidad: unidad, lector: lector}
}

func (caso *DetalleDeMina) Ejecutar(ctx context.Context, identificadorMina string) (puertos.DetalleMina, bool, error) {
	var detalle puertos.DetalleMina
	var encontrada bool
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		resultado, existe, err := caso.lector.DetalleDeMina(ctx, identificadorMina)
		detalle, encontrada = resultado, existe
		return err
	})
	return detalle, encontrada, err
}
