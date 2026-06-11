package aplicacion

import (
	"context"

	"minas/capacidades/catalogos/puertos"
)

type DetalleDeEquipo struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeCatalogos
}

func NuevoDetalleDeEquipo(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeCatalogos) *DetalleDeEquipo {
	return &DetalleDeEquipo{unidad: unidad, lector: lector}
}

func (caso *DetalleDeEquipo) Ejecutar(ctx context.Context, identificadorEquipo string) (puertos.DetalleEquipo, bool, error) {
	var detalle puertos.DetalleEquipo
	var encontrado bool
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		resultado, existe, err := caso.lector.DetalleDeEquipo(ctx, identificadorEquipo)
		detalle, encontrado = resultado, existe
		return err
	})
	return detalle, encontrado, err
}
