package aplicacion

import (
	"context"

	"minas/capacidades/catalogos/puertos"
)

type DetalleDeEmpleado struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeCatalogos
}

func NuevoDetalleDeEmpleado(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeCatalogos) *DetalleDeEmpleado {
	return &DetalleDeEmpleado{unidad: unidad, lector: lector}
}

func (caso *DetalleDeEmpleado) Ejecutar(ctx context.Context, identificadorEmpleado string) (puertos.DetalleEmpleado, bool, error) {
	var detalle puertos.DetalleEmpleado
	var encontrado bool
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		resultado, existe, err := caso.lector.DetalleDeEmpleado(ctx, identificadorEmpleado)
		detalle, encontrado = resultado, existe
		return err
	})
	return detalle, encontrado, err
}
