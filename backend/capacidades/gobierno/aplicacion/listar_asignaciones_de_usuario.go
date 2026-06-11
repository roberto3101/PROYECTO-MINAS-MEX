package aplicacion

import (
	"context"

	"minas/capacidades/gobierno/puertos"
	"minas/compartido/identificador"
)

type ListarAsignacionesDeUsuario struct {
	unidad puertos.UnidadDeTrabajo
	lector puertos.LectorDeGobierno
}

func NuevoListarAsignacionesDeUsuario(unidad puertos.UnidadDeTrabajo, lector puertos.LectorDeGobierno) *ListarAsignacionesDeUsuario {
	return &ListarAsignacionesDeUsuario{unidad: unidad, lector: lector}
}

func (caso *ListarAsignacionesDeUsuario) Ejecutar(ctx context.Context, identificadorUsuario string) ([]puertos.ResumenAsignacion, error) {
	usuario, err := identificador.Desde(identificadorUsuario)
	if err != nil {
		return nil, err
	}
	var asignaciones []puertos.ResumenAsignacion
	err = caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		listadas, err := caso.lector.AsignacionesDe(ctx, usuario.Texto())
		asignaciones = listadas
		return err
	})
	return asignaciones, err
}
