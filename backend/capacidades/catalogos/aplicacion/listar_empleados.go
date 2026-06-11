package aplicacion

import (
	"context"

	"minas/capacidades/catalogos/puertos"
)

type ListarEmpleados struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeCatalogos
}

func NuevoListarEmpleados(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeCatalogos) *ListarEmpleados {
	return &ListarEmpleados{unidad: unidad, lector: lector}
}

func (caso *ListarEmpleados) Ejecutar(ctx context.Context) ([]puertos.ResumenEmpleado, error) {
	var empleados []puertos.ResumenEmpleado
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listados, err := caso.lector.ListarEmpleados(ctx)
		empleados = listados
		return err
	})
	return empleados, err
}
