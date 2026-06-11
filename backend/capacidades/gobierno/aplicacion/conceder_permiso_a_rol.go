package aplicacion

import (
	"context"

	"minas/capacidades/gobierno/puertos"
	"minas/compartido/identificador"
)

type ComandoConcederPermisoARol struct {
	IdentificadorRol string
	CodigoPermiso    string
}

type ConcederPermisoARol struct {
	unidad   puertos.UnidadDeTrabajo
	roles    puertos.RepositorioRol
	permisos puertos.RepositorioPermiso
}

func NuevoConcederPermisoARol(unidad puertos.UnidadDeTrabajo, roles puertos.RepositorioRol, permisos puertos.RepositorioPermiso) *ConcederPermisoARol {
	return &ConcederPermisoARol{unidad: unidad, roles: roles, permisos: permisos}
}

func (caso *ConcederPermisoARol) Ejecutar(ctx context.Context, comando ComandoConcederPermisoARol) error {
	identificadorRol, err := identificador.Desde(comando.IdentificadorRol)
	if err != nil {
		return err
	}
	return caso.unidad.EnTransaccion(ctx, func(ctx context.Context) error {
		rol, encontrado, err := caso.roles.BuscarPorIdentificador(ctx, identificadorRol)
		if err != nil {
			return err
		}
		if !encontrado {
			return ErrRolNoEncontrado
		}
		permiso, existe, err := caso.permisos.BuscarPorCodigo(ctx, comando.CodigoPermiso)
		if err != nil {
			return err
		}
		if !existe {
			return ErrPermisoNoEncontrado
		}
		if err := rol.ConcederPermiso(permiso.Identificador()); err != nil {
			return err
		}
		return caso.roles.Guardar(ctx, rol)
	})
}
