package aplicacion

import (
	"context"

	"minas/capacidades/gobierno/puertos"
)

type ListarUsuarios struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeGobierno
}

func NuevoListarUsuarios(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeGobierno) *ListarUsuarios {
	return &ListarUsuarios{unidad: unidad, lector: lector}
}

func (caso *ListarUsuarios) Ejecutar(ctx context.Context) ([]puertos.ResumenUsuario, error) {
	var usuarios []puertos.ResumenUsuario
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listados, err := caso.lector.ListarUsuarios(ctx)
		usuarios = listados
		return err
	})
	return usuarios, err
}
