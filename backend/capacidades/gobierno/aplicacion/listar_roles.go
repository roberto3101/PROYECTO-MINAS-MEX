package aplicacion

import (
	"context"

	"minas/capacidades/gobierno/puertos"
)

type ListarRoles struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeGobierno
}

func NuevoListarRoles(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeGobierno) *ListarRoles {
	return &ListarRoles{unidad: unidad, lector: lector}
}

func (caso *ListarRoles) Ejecutar(ctx context.Context) ([]puertos.ResumenRol, error) {
	var roles []puertos.ResumenRol
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listados, err := caso.lector.ListarRoles(ctx)
		roles = listados
		return err
	})
	return roles, err
}
