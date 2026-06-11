package aplicacion

import (
	"context"

	"minas/capacidades/gobierno/puertos"
)

type ListarPermisos struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeGobierno
}

func NuevoListarPermisos(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeGobierno) *ListarPermisos {
	return &ListarPermisos{unidad: unidad, lector: lector}
}

func (caso *ListarPermisos) Ejecutar(ctx context.Context) ([]puertos.ResumenPermiso, error) {
	var permisos []puertos.ResumenPermiso
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listados, err := caso.lector.CatalogoDePermisos(ctx)
		permisos = listados
		return err
	})
	return permisos, err
}
