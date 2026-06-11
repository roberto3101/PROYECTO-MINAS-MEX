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

func (caso *ListarEmpleados) Ejecutar(ctx context.Context, filtro puertos.FiltroDeCatalogo) ([]puertos.ResumenEmpleado, string, error) {
	var empleados []puertos.ResumenEmpleado
	var siguiente string
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listados, cursor, err := caso.lector.ListarEmpleados(ctx, filtro)
		empleados, siguiente = listados, cursor
		return err
	})
	return empleados, siguiente, err
}
