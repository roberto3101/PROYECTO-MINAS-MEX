package aplicacion

import (
	"context"

	"minas/capacidades/gobierno/puertos"
)

type MinasDeSesion struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeGobierno
}

func NuevoMinasDeSesion(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeGobierno) *MinasDeSesion {
	return &MinasDeSesion{unidad: unidad, lector: lector}
}

func (caso *MinasDeSesion) Ejecutar(ctx context.Context, identificadorUsuario string) (puertos.AccesoAMinas, error) {
	var acceso puertos.AccesoAMinas
	err := caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		resultado, err := caso.lector.MinasAccesibles(ctx, identificadorUsuario)
		acceso = resultado
		return err
	})
	return acceso, err
}
