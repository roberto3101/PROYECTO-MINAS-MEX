package aplicacion

import (
	"context"

	"minas/capacidades/gobierno/puertos"
	"minas/compartido/identificador"
)

type ComandoRevocarRol struct {
	IdentificadorAsignacion string
}

type RevocarRol struct {
	unidad       puertos.UnidadDeTrabajo
	asignaciones puertos.RepositorioAsignacionRol
}

func NuevoRevocarRol(unidad puertos.UnidadDeTrabajo, asignaciones puertos.RepositorioAsignacionRol) *RevocarRol {
	return &RevocarRol{unidad: unidad, asignaciones: asignaciones}
}

func (caso *RevocarRol) Ejecutar(ctx context.Context, comando ComandoRevocarRol) error {
	identificadorAsignacion, err := identificador.Desde(comando.IdentificadorAsignacion)
	if err != nil {
		return err
	}
	return caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		asignacion, encontrada, err := caso.asignaciones.BuscarPorIdentificador(ctx, identificadorAsignacion)
		if err != nil {
			return err
		}
		if !encontrada {
			return ErrAsignacionNoEncontrada
		}
		if err := asignacion.Revocar(); err != nil {
			return err
		}
		return caso.asignaciones.Guardar(ctx, asignacion)
	})
}
